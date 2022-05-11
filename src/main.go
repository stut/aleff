package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-acme/lego/v4/log"
)

var (
	metricRunCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "aleff_run_total",
		Help: "The total number of runs",
	})
	metricLastRunTimestamp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "aleff_last_run_timestamp_seconds",
		Help: "The unix timestamp when the last run was started",
	}, []string{
		"dir_url",
		"email",
		"domain",
	})
	metricDomainCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "aleff_domain_total",
		Help: "The total number of domains detected",
	})
	metricErrorCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "aleff_errors_total",
		Help: "The total number of errors",
	}, []string{
		"dir_url",
		"email",
		"domain",
		"reason",
	})
	metricCertificateObtainedTimestamp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "aleff_certificate_obtained_timestamp_seconds",
		Help: "The unix timestamp of the last time a certificate was obtained",
	}, []string{
		"dir_url",
		"email",
		"domain",
	})
	metricCertificateExpiryTimestamp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "aleff_certificate_expiry_timestamp_seconds",
		Help: "The unix timestamp when the certificate will expire",
	}, []string{
		"dir_url",
		"email",
		"domain",
	})
)

func env(name string, required bool) string {
	val := os.Getenv(name)
	if len(val) == 0 && required {
		metricErrorCounter.WithLabelValues("-", "-", "-", fmt.Sprintf("env-misssing-%s", name))
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
		metricErrorCounter.WithLabelValues("-", "-", "-", "duration-parse-failure-RUN_INTERVAL")
		log.Fatalf("Failed to parse RUN_INTERVAL: %v", err)
	}

	// Default renewWithin is 29 days.
	renewWithinDefault := fmt.Sprintf("%dh", 24*29)
	renewWithin, err = time.ParseDuration(envOrDefault("RENEW_WITHIN", renewWithinDefault))
	if err != nil {
		metricErrorCounter.WithLabelValues("-", "-", "-", "duration-parse-failure-RENEW_WITHIN")
		log.Fatalf("Failed to parse RENEW_WITHIN: %v", err)
	}

	metricsListenAddress := envOrDefault("PROMETHEUS_LISTEN_ADDRESS", ":2123")

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

	// Start the prometheus metrics service.
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Infof("aleff: Prometheus metrics available at http://%s/metrics", metricsListenAddress)
		log.Fatal(http.ListenAndServe(metricsListenAddress, nil))
	}()

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
