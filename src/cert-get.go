package main

import (
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

func (manager *Manager) getCertificate(domain string) error {
	privateKey, err := manager.getPrivateKey()
	if err != nil {
		return err
	}

	myUser, err := manager.getUser(privateKey)

	config := lego.NewConfig(myUser)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	config.CADirURL = manager.acmeDirectoryUrl
	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		return err
	}

	err = client.Challenge.SetHTTP01Provider(manager)
	if err != nil {
		return err
	}

	// New users will need to register
	if myUser.Registration == nil {
		reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return err
		}
		myUser.Registration = reg
		err = manager.saveUser(myUser)
		if err != nil {
			return err
		}
	}

	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  false,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return err
	}

	err = manager.setValueInConsul(manager.getCertificateKey(domain, "key"), certificates.PrivateKey)
	if err != nil {
		return err
	}

	err = manager.setValueInConsul(manager.getCertificateKey(domain, "cert"), certificates.Certificate)
	if err != nil {
		return err
	}

	return nil
}
