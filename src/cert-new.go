package main

func (manager *Manager) newCertificate(domain string) error {
	return manager.getCertificate(domain)
}
