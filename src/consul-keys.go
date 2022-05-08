package main

import "fmt"

func (manager *Manager) getPrivateKeyKey() string {
	return fmt.Sprintf("%s%s/%s", manager.kvConfigRoot, manager.acmeDirectoryUrl, "private-key")
}

func (manager *Manager) getUserKey() string {
	return fmt.Sprintf("%s%s/%s", manager.kvConfigRoot, manager.acmeDirectoryUrl, "user")
}

func (manager *Manager) getChallengeKey(token string) string {
	return fmt.Sprintf("%s%s", manager.kvChallengeRoot, token)
}

func (manager *Manager) getCertificateKey(domain, part string) string {
	return fmt.Sprintf("%s%s-%s.pem", manager.kvCertRoot, domain, part)
}
