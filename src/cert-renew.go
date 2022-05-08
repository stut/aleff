package main

func (manager *Manager) renewCertificate(domain string) error {
	return manager.getCertificate(domain)
}
