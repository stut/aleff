package main

import "log"

func (manager *Manager) renewCertificate(domain string) error {
	log.Printf("Renewing certificate for %s", domain)
	return nil
}
