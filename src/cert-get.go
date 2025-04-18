package main

import (
	"fmt"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

func (manager *Manager) getCertificate(domain string) error {
	privateKey, err := manager.getPrivateKey()
	if err != nil {
		manager.markErrorMetric(domain, "fetch-private-key")
		return fmt.Errorf("failed to fetch or create private key: %v", err)
	}

	var myUser *MyUser
	myUser, err = manager.getUser(privateKey)
	if err != nil {
		manager.markErrorMetric(domain, "fetch-user")
		return fmt.Errorf("failed to fetch or create user object: %v", err)
	}

	config := lego.NewConfig(myUser)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	config.CADirURL = manager.acmeDirectoryUrl
	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	var client *lego.Client
	client, err = lego.NewClient(config)
	if err != nil {
		manager.markErrorMetric(domain, "lego-client")
		return fmt.Errorf("failed to create lego client: %v", err)
	}

	err = client.Challenge.SetHTTP01Provider(manager)
	if err != nil {
		manager.markErrorMetric(domain, "http01-provider")
		return fmt.Errorf("failed to create HTTP01 provider: %v", err)
	}

	// New users will need to register
	if myUser.Registration == nil {
		reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			manager.markErrorMetric(domain, "register-user")
			return fmt.Errorf("failed to register user: %v", err)
		}
		myUser.Registration = reg
		err = manager.saveUser(myUser)
		if err != nil {
			manager.markErrorMetric(domain, "store-user")
			return fmt.Errorf("failed to store user object: %v", err)
		}
	}

	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		manager.markErrorMetric(domain, "obtain-certificate")
		return fmt.Errorf("failed to obtain certificate: %v", err)
	}

	metricCertificateObtainedTimestamp.WithLabelValues(manager.acmeDirectoryUrl, manager.emailAddress, domain).SetToCurrentTime()

	err = manager.setValueInConsul(manager.getCertificateKey(domain, "key"), certificates.PrivateKey)
	if err != nil {
		manager.markErrorMetric(domain, "store-certificate-key")
		return fmt.Errorf("failed to store certificate key: %v", err)
	}

	err = manager.setValueInConsul(manager.getCertificateKey(domain, "cert"), certificates.Certificate)
	if err != nil {
		manager.markErrorMetric(domain, "store-certificate-cert")
		return fmt.Errorf("failed to store certificate cert: %v", err)
	}

	return nil
}
