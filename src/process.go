package main

import (
	"crypto/x509"
	"log"
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

	timeNow := time.Now()

	if leafCert == nil {
		log.Printf("    Requesting new...")
		err = manager.newCertificate(domain)
	} else if timeNow.Add(manager.renewWithin).After(leafCert.NotAfter) {
		log.Printf("    Renewing...")
		err = manager.renewCertificate(domain)
	} else {
		log.Printf("    Valid, expires %v", leafCert.NotAfter)
		return nil
	}

	return err
}
