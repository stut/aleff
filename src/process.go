package main

import (
	"crypto/x509"
	"fmt"
	"github.com/go-acme/lego/v4/log"
	"time"
)

func (manager *Manager) processDomain(domain string) error {
	var leafCert *x509.Certificate

	cert, err := manager.GetCertificateFromConsul(domain)
	if err != nil {
		manager.markErrorMetric(domain, "fetch-existing-certificate")
		return fmt.Errorf("failed to fetch existing certificate: %v", err)
	}

	if cert != nil {
		leafCert, err = x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			manager.markErrorMetric(domain, "parse-existing-certificate")
			log.Warnf("[%s] aleff: Failed to parse existing certificate: %v", domain, err)
			// Force requesting a new certificate.
			leafCert = nil
		}
	}

	if leafCert == nil {
		log.Infof("[%s] aleff: Requesting new certificate...", domain)
		err = manager.newCertificate(domain)
	} else {
		metricCertificateExpiryTimestamp.WithLabelValues(manager.acmeDirectoryUrl, manager.emailAddress, domain).Set(float64(leafCert.NotAfter.Unix()))
		hoursToRenewal := int(leafCert.NotAfter.Sub(time.Now().UTC()).Hours() - manager.renewWithin.Hours())

		if hoursToRenewal <= 0 {
			log.Infof("[%s] aleff: Renewing certificate...", domain)
			err = manager.renewCertificate(domain)
		} else {
			log.Infof("[%s] aleff: Renewal due in %d hours", domain, hoursToRenewal)
			err = nil
		}
	}

	return err
}
