package main

import (
	"errors"
	"github.com/go-acme/lego/v4/log"
	consulApi "github.com/hashicorp/consul/api"
	nomadApi "github.com/hashicorp/nomad/api"
	"golang.org/x/crypto/acme"
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
	acmeDirectoryUrl              string
	renewWithin                   time.Duration
	challengeResponderJobFilename string
	challengeResponderJob         *nomadApi.Job
}

func createManager(emailAddress string, tagPrefix string, configRoot string, certRoot string, challengeRoot string, acmeDirectoryUrl string, renewWithin time.Duration, challengeResponderJobFilename string) *Manager {
	var err error

	// Make sure the challenge responder job definition file exists.
	if _, err := os.Stat(challengeResponderJobFilename); errors.Is(err, os.ErrNotExist) {
		metricErrorCounter.WithLabelValues(acmeDirectoryUrl, emailAddress, "-", "missing-challenge-responder-job-file")
		log.Fatalf("Challenge responder job definition file not found: %s", challengeResponderJobFilename)
	}

	manager := &Manager{
		emailAddress:                  emailAddress,
		tagPrefix:                     tagPrefix,
		kvConfigRoot:                  configRoot,
		kvCertRoot:                    certRoot,
		kvChallengeRoot:               challengeRoot,
		acmeDirectoryUrl:              acmeDirectoryUrl,
		renewWithin:                   renewWithin,
		challengeResponderJobFilename: challengeResponderJobFilename,
	}

	manager.consulClient, err = consulApi.NewClient(consulApi.DefaultConfig())
	if err != nil {
		manager.markErrorMetric("-", "cannot-create-consul-client")
		log.Fatalf("Failed to create consul client: %v", err)
	}

	manager.consulKv = manager.consulClient.KV()

	return manager
}

func (manager *Manager) markErrorMetric(domain, reason string) {
	metricErrorCounter.WithLabelValues(manager.acmeDirectoryUrl, manager.emailAddress, domain, reason)
}

func (manager *Manager) run() {
	log.Infof("aleff: Running...")
	metricRunCounter.Inc()
	domains, err := manager.discoverDomainsFromConsul()
	if err != nil {
		manager.markErrorMetric("-", "discover-domains")
		log.Warnf("aleff: Failed to discover domains from Consul: %v", err)
		return
	}
	metricDomainCount.Set(float64(len(domains)))

	for _, domain := range domains {
		metricLastRunTimestamp.WithLabelValues(manager.acmeDirectoryUrl, manager.emailAddress, domain).SetToCurrentTime()
		err = manager.processDomain(domain)
		if err != nil {
			// No error metric recorded here as it will have been recorded with a correct reason label where the error
			// occurred.
			log.Warnf("[%s] aleff: Error: %v", domain, err)
		}
	}
	log.Infof("aleff: Done.")
}
