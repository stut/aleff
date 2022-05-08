package main

import (
	"crypto/x509"
	"github.com/go-acme/lego/v4/log"
	"time"
)

func (manager *Manager) processDomain(domain string) error {
	var leafCert *x509.Certificate

	cert, err := manager.GetCertificateFromConsul(domain)
	if err != nil {
		return err
	}

	if cert != nil {
		leafCert, err = x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return err
		}
	}

	if leafCert == nil {
		log.Infof("[%s] aleff: Requesting new certificate...", domain)
		err = manager.newCertificate(domain)
	} else {
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
