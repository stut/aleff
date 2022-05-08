package main

import (
	"fmt"
	"github.com/go-acme/lego/v4/log"
	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/jobspec"
	"io/ioutil"
	"net/http"
	"time"
)

func (manager *Manager) Present(domain, token, keyAuth string) error {
	err := manager.setValueInConsul(manager.getChallengeKey(token), []byte(keyAuth))
	if err != nil {
		return err
	}

	err = manager.startChallengeResponder(domain, token, keyAuth)
	if err != nil {
		return err
	}
	return nil
}

func (manager *Manager) CleanUp(domain, token, keyAuth string) error {
	err := manager.DeleteValueInConsul(manager.getChallengeKey(token))
	if err != nil {
		return err
	}

	err = manager.stopChallengeResponder(domain)
	if err != nil {
		return err
	}

	return nil
}

func (manager *Manager) startChallengeResponder(domain, token, keyAuth string) error {
	log.Infof("[%s] aleff: Starting challenge responder...", domain)
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err
	}

	manager.challengeResponderJob, err = jobspec.ParseFile(manager.challengeResponderJobFilename)
	manager.challengeResponderJob.TaskGroups[0].Tasks[0].Services[0].Tags = append(
		manager.challengeResponderJob.TaskGroups[0].Tasks[0].Services[0].Tags,
		fmt.Sprintf("urlprefix-%s/.well-known/acme-challenge/", domain))

	_, _, err = client.Jobs().Register(manager.challengeResponderJob, nil)
	if err != nil {
		return err
	}

	// Give the job 60 seconds (12 tries with a 5 second delay) to respond correctly.
	url := fmt.Sprintf("http://%s/.well-known/acme-challenge/%s", domain, token)
	for i := 0; i < 12; i++ {
		resp, err := http.Get(url)
		if err != nil {
			log.Warnf("[%s] aleff: Challenge responder still starting up...", domain)
		} else {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Warnf("[%s] aleff: Challenge responder still spinning up...", domain)
			} else {
				bodyString := string(body)
				if bodyString != keyAuth {
					log.Warnf("[%s] aleff: Challenge responder job still spinning up...", domain)
				} else {
					// Challenge responder is responding correctly, continue the validation process.
					log.Infof("[%s] aleff: Challenge responder started, waiting for challenge...", domain)
					return nil
				}
			}
		}
		time.Sleep(time.Second * 5)
	}

	return fmt.Errorf("challenge responder did not start correctly within 30 seconds")
}

func (manager *Manager) stopChallengeResponder(domain string) error {
	log.Infof("[%s] aleff: Stopping challenge responder...", domain)
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
