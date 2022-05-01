package main

import (
	"crypto"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

// User object that implements acme.User
type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func (manager *Manager) newCertificate(domain string) error {
	privateKey, err := manager.getPrivateKey()
	if err != nil {
		return err
	}

	myUser := MyUser{
		Email: manager.emailAddress,
		key:   privateKey,
	}

	config := lego.NewConfig(&myUser)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	config.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
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
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return err
	}
	myUser.Registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  false,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return err
	}

	err = manager.SetValueInConsul(manager.GetCertificateKey(domain, "key"), certificates.PrivateKey)
	if err != nil {
		return err
	}

	err = manager.SetValueInConsul(manager.GetCertificateKey(domain, "cert"), certificates.Certificate)
	if err != nil {
		return err
	}

	return nil
}
