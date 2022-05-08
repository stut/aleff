package main

import (
	"fmt"
	"strings"
)

func (manager *Manager) directoryUrlToKey() string {
	return strings.ReplaceAll(manager.acmeDirectoryUrl, "/", "-")
}

func (manager *Manager) getPrivateKeyKey() string {
	return fmt.Sprintf("%s%s/%s", manager.kvConfigRoot, manager.directoryUrlToKey(), "private-key")
}

func (manager *Manager) getUserKey() string {
	return fmt.Sprintf("%s%s/%s", manager.kvConfigRoot, manager.directoryUrlToKey(), "user")
}

func (manager *Manager) getChallengeKey(token string) string {
	return fmt.Sprintf("%s%s", manager.kvChallengeRoot, token)
}

func (manager *Manager) getCertificateKey(domain, part string) string {
	return fmt.Sprintf("%s%s-%s.pem", manager.kvCertRoot, domain, part)
}
