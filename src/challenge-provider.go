package main

import (
	"fmt"
	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/jobspec"
	"log"
	"time"
)

func (manager *Manager) Present(domain, token, keyAuth string) error {
	err := manager.StartChallengeResponder(domain)
	if err != nil {
		return err
	}
	return manager.SetValueInConsul(manager.GetChallengeKey(token), []byte(keyAuth))
}

func (manager *Manager) CleanUp(domain, token, keyAuth string) error {
	err := manager.StopChallengeResponder()
	if err != nil {
		return err
	}
	return manager.DeleteValueInConsul(manager.GetChallengeKey(token))
}

func (manager *Manager) StartChallengeResponder(domain string) error {
	log.Printf("Starting challenge responder...")
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err
	}

	manager.challengeResponderJob, err = jobspec.ParseFile(manager.challengeResponderJobFilename)
	manager.challengeResponderJob.TaskGroups[0].Tasks[0].Services[0].Tags = append(
		manager.challengeResponderJob.TaskGroups[0].Tasks[0].Services[0].Tags,
		fmt.Sprintf("urlprefix-%s/.well-known/acme-challenge/", domain))

	_, _, err = client.Jobs().Register(manager.challengeResponderJob, nil)

	// TODO: Is there a way to detect when Fabio has picked up the route rather than waiting so long?
	time.Sleep(time.Second * 30)

	return err
}

func (manager *Manager) StopChallengeResponder() error {
	log.Printf("Stopping challenge responder...")
	if manager.challengeResponderJob == nil {
		return fmt.Errorf("")
	}

	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err
	}

	stopIt := true
	manager.challengeResponderJob.Stop = &stopIt

	_, _, err = client.Jobs().Register(manager.challengeResponderJob, nil)
	return err
}
