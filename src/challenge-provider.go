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

	if !manager.disableChallengeResponderJob {
		err = manager.startChallengeResponder(domain, token, keyAuth)
		if err != nil {
			return err
		}
	}

	return nil
}

func (manager *Manager) CleanUp(domain, token, keyAuth string) error {
	err := manager.DeleteValueInConsul(manager.getChallengeKey(token))
	if err != nil {
		return err
	}

	if !manager.disableChallengeResponderJob {
		err = manager.stopChallengeResponder(domain)
		if err != nil {
			return err
		}
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
		fmt.Sprintf("%s%s:80/.well-known/acme-challenge/", manager.tagPrefix, domain))

	_, _, err = client.Jobs().Register(manager.challengeResponderJob, nil)
	if err != nil {
		return err
	}

	// Give the job the configured amount of time to respond correctly.
	timeoutTime := time.Now().Add(manager.challengeResponderJobTimeout)
	url := fmt.Sprintf("http://%s/.well-known/acme-challenge/%s", domain, token)
	for {
		if time.Now().After(timeoutTime) {
			break
		}

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

	// Ignore any errors stopping the job if we've timed out. If it fails it has no effect other than being untidy.
	_ = manager.stopChallengeResponder(domain)
	return fmt.Errorf("challenge responder did not start correctly within %s", manager.challengeResponderJobTimeout.String())
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
