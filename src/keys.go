package main

import "fmt"

func (manager *Manager) GetChallengeKey(token string) string {
	return fmt.Sprintf("%s%s", manager.kvChallengeRoot, token)
}

func (manager *Manager) GetCertificateKey(domain, part string) string {
	return fmt.Sprintf("%s%s-%s.pem", manager.kvCertRoot, domain, part)
}
