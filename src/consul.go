package main

import (
	"crypto/tls"
	"github.com/hashicorp/consul/api"
	"sort"
	"strings"
)

func (manager *Manager) discoverDomainsFromConsul() ([]string, error) {
	catalog := manager.consulClient.Catalog()
	services, _, err := catalog.Services(nil)
	if err != nil {
		return nil, err
	}

	domainMap := make(map[string]bool)
	for _, tags := range services {
		domains := []string{}
		enabled := false
		disabled := false
		for _, tag := range tags {
			// Exclude any challenge responder tags that might have been left around.
			if strings.Contains(tag, ".well-known/acme-challenge") {
				continue
			}

			if strings.HasPrefix(tag, manager.tagPrefix) {
				url := strings.SplitN(tag, manager.tagPrefix, 2)[1]
				url = strings.SplitN(url, " ", 2)[0]
				if url[0] != '/' {
					separator := "/"
					if strings.Contains(url, ":") {
						separator = ":"
					}
					domains = append(domains, strings.SplitN(url, separator, 2)[0])
				}
			} else if tag == "aleff-enabled" {
				enabled = true
			} else if tag == "aleff-disabled" {
				disabled = true
			}
		}

		if len(domains) > 0 && (manager.defaultEnabled && !disabled) || (!manager.defaultEnabled && enabled) {
			for _, domain := range domains {
				domainMap[domain] = true
			}
		}
	}

	var domains []string
	for domain, _ := range domainMap {
		domains = append(domains, domain)
	}
	sort.Strings(domains)
	return domains, nil
}

func (manager *Manager) getValueFromConsul(key string) (value []byte, err error) {
	var pair *api.KVPair
	pair, _, err = manager.consulKv.Get(key, nil)
	if err != nil {
		return nil, err
	}

	if pair == nil {
		return nil, nil
	}

	return pair.Value, nil
}

func (manager *Manager) setValueInConsul(key string, value []byte) error {
	pair := &api.KVPair{
		Key:   key,
		Value: value,
	}

	_, err := manager.consulKv.Put(pair, nil)

	if err != nil {
		return err
	}

	return nil
}

func (manager *Manager) DeleteValueInConsul(key string) error {
	_, err := manager.consulKv.Delete(key, nil)
	return err
}

func (manager *Manager) GetCertificateFromConsul(domain string) (*tls.Certificate, error) {
	certBytes, err := manager.getValueFromConsul(manager.getCertificateKey(domain, "cert"))
	if err != nil {
		return nil, err
	}

	if certBytes == nil {
		return nil, nil
	}

	keyBytes, err := manager.getValueFromConsul(manager.getCertificateKey(domain, "key"))
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return nil, err
	}

	return &cert, nil
}
