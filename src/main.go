package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func envOrDefault(name string, defaultValue string) string {
	val := os.Getenv(name)
	if len(val) == 0 {
		return defaultValue
	}
	return val
}

func main() {
	var runInterval, renewWithin time.Duration
	var err error

	// A RUN_INTERVAL of 0 will run just once and exit.
	runInterval, err = time.ParseDuration(envOrDefault("RUN_INTERVAL", "0"))
	if err != nil {
		panic(err)
	}

	// Default renewWithin is 29 days.
	renewWithin, err = time.ParseDuration(envOrDefault("RENEW_WITHIN", fmt.Sprintf("%dh", 24*29)))
	if err != nil {
		panic(err)
	}

	manager := createManager(
		envOrDefault("EMAIL_ADDRESS", "stuart@stut.net"),
		envOrDefault("TAG_PREFIX", "urlprefix-"),
		envOrDefault("KV_STATUS_ROOT", "certs/status/"),
		envOrDefault("KV_CONFIG_ROOT", "certs/config/"),
		envOrDefault("KV_CERT_ROOT", "certs/active/"),
		envOrDefault("KV_CHALLENGE_ROOT", "certs/challenges/"),
		renewWithin,
		envOrDefault("CHALLENGE_RESPONDER_JOB_FILENAME", "local/challenge-responder.hcl"))

	if runInterval.Seconds() == 0 {
		manager.run()
		os.Exit(0)
	}

	// Catch signals so the current run can finish rather than killing it mid-run.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)

	for {
		manager.run()

		select {
		case <-sigs:
			os.Exit(1)
		case <-time.After(runInterval):
			break
		}
	}
}
