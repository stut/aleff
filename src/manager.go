package main

import (
	"errors"
	consulApi "github.com/hashicorp/consul/api"
	nomadApi "github.com/hashicorp/nomad/api"
	"golang.org/x/crypto/acme"
	"log"
	"os"
	"time"
)

type Manager struct {
	emailAddress                  string
	tagPrefix                     string
	kvConfigRoot                  string
	kvCertRoot                    string
	kvChallengeRoot               string
	consulClient                  *consulApi.Client
	consulKv                      *consulApi.KV
	client                        *acme.Client
	renewWithin                   time.Duration
	challengeResponderJobFilename string
	challengeResponderJob         *nomadApi.Job
}

func createManager(emailAddress string, tagPrefix string, configRoot string, certRoot string, challengeRoot string, renewWithin time.Duration, challengeResponderJobFilename string) *Manager {
	var err error

	// Make sure the challenge responder job definition file exists.
	if _, err := os.Stat(challengeResponderJobFilename); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("Challenge responder job definition file not found: %s", challengeResponderJobFilename)
	}

	manager := &Manager{
		emailAddress:                  emailAddress,
		tagPrefix:                     tagPrefix,
		kvConfigRoot:                  configRoot,
		kvCertRoot:                    certRoot,
		kvChallengeRoot:               challengeRoot,
		renewWithin:                   renewWithin,
		challengeResponderJobFilename: challengeResponderJobFilename,
	}

	manager.consulClient, err = consulApi.NewClient(consulApi.DefaultConfig())
	if err != nil {
		log.Fatalf("Failed to create consul client: %v", err)
	}

	manager.consulKv = manager.consulClient.KV()

	return manager
}

func (manager *Manager) run() {
	log.Printf("Processing...")
	domains, err := manager.DiscoverDomainsFromConsul()
	if err != nil {
		panic(err)
	}

	for _, domain := range domains {
		log.Printf("  %s...", domain)
		err = manager.processDomain(domain)
		if err != nil {
			log.Printf("    Error processing %s: %v\n", domain, err)
		}
	}
	log.Printf("Done.")
}
