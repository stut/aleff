package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-acme/lego/v4/log"
)

func env(name string, required bool) string {
	val := os.Getenv(name)
	if len(val) == 0 && required {
		log.Fatalf("Environment variable %s is required!", name)
	}
	return val
}

func envOrDefault(name string, defaultValue string) string {
	val := env(name, false)
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
		log.Fatalf("Failed to parse RUN_INTERVAL: %v", err)
	}

	// Default renewWithin is 29 days.
	renewWithinDefault := fmt.Sprintf("%dh", 24*29)
	renewWithin, err = time.ParseDuration(envOrDefault("RENEW_WITHIN", renewWithinDefault))
	if err != nil {
		log.Fatalf("Failed to parse RENEW_WITHIN: %v", err)
	}

	manager := createManager(
		env("EMAIL_ADDRESS", true),
		envOrDefault("TAG_PREFIX", "urlprefix-"),
		envOrDefault("KV_CONFIG_ROOT", "certs/config/"),
		envOrDefault("KV_CERT_ROOT", "certs/active/"),
		envOrDefault("KV_CHALLENGE_ROOT", "certs/challenges/"),
		envOrDefault("ACME_DIR_URL", "https://acme-v02.api.letsencrypt.org/directory"),
		renewWithin,
		env("CHALLENGE_RESPONDER_JOB_FILENAME", true))

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
			os.Exit(0)
		case <-time.After(runInterval):
			break
		}
	}
}
