package main

import (
	"crypto"
	"crypto/ecdsa"
	"encoding/json"
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

func (manager *Manager) getUser(privateKey *ecdsa.PrivateKey) (*MyUser, error) {
	return manager.loadUser(privateKey)
}

func (manager *Manager) loadUser(privateKey *ecdsa.PrivateKey) (*MyUser, error) {
	var err error
	var myUser *MyUser

	userKey := manager.getUserKey()
	userBytes, _ := manager.getValueFromConsul(userKey)
	if userBytes == nil {
		// Create a new user object.
		myUser = &MyUser{
			Email: manager.emailAddress,
			key:   privateKey,
		}
		err := manager.saveUser(myUser)
		if err != nil {
			return nil, err
		}
	} else {
		err = json.Unmarshal(userBytes, &myUser)
		if err != nil {
			return nil, err
		}
		myUser.key = privateKey
	}

	return myUser, nil
}

func (manager *Manager) saveUser(myUser *MyUser) error {
	userBytes, err := json.Marshal(myUser)
	if err != nil {
		return err
	}
	return manager.setValueInConsul(manager.getUserKey(), userBytes)
}
