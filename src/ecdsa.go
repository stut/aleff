package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
)

func (manager *Manager) getPrivateKey() (*ecdsa.PrivateKey, error) {
	consulKey := manager.getPrivateKeyKey()
	privateKeyBytes, _ := manager.getValueFromConsul(consulKey)

	if privateKeyBytes == nil {
		// Generate and store a new key.
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}
		privateKeyBytes = encodeKey(privateKey)
		err = manager.setValueInConsul(consulKey, privateKeyBytes)
	}

	return decodeKey(privateKeyBytes)
}

func encodeKey(privateKey *ecdsa.PrivateKey) []byte {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})
}

func decodeKey(pemEncoded []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemEncoded)
	x509Encoded := block.Bytes
	return x509.ParseECPrivateKey(x509Encoded)
}
