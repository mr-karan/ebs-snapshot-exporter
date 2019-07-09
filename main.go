/*
ebs-snapshot-exporter is used to fetch metadata about EBS Snapshots and exports
them as Prometheus metrics. EBS Snapshots are used to take incremental backups of
AWS EBS Volumes. AWS Users can automate the snapshots using Amazon DLM (Data Lifecycle Manager).
ebs-snapshot-exporter exports metrics such as the
 - size of EBS volume
 - total count of snapshots
 - timestamp of snapshot created

Usage Instructions
`./ebs-snapshot-exporter --config=config.yml`
*/

package main

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// injected during build
	buildVersion = "unknown"
	buildDate    = "unknown"
)

func initLogger(config cfgApp) *logrus.Logger {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	// Set logger level
	switch level := config.LogLevel; level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
		logger.Debug("verbose logging enabled")
	default:
		logger.SetLevel(logrus.InfoLevel)
	}
	return logger
}

func main() {
	var (
		config = initConfig()
		logger = initLogger(config.App)
	)
	// Initialize hub.
	hub := &Hub{
		config:  config,
		logger:  logger,
		version: buildVersion,
	}
	hub.logger.Infof("booting ebs-snapshot-exporter-version:%v", buildVersion)
	// Initialize prometheus registry.
	r := prometheus.NewRegistry()
	// Fetch all jobs listed in config.
	for _, job := range hub.config.App.Jobs {
		// Initialize the exporter. Exporter is a collection of metrics to be exported.
		exporter, err := hub.NewExporter(&job)
		if err != nil {
			hub.logger.Errorf("exporter initialization failed for %s", job.Name)
		}
		// Register the exporters with our custom registry. Panics in case of failure.
		r.MustRegister(exporter)
		hub.logger.Debugf("registration of metrics for job %s success", job.Name)
	}
	// Default index handler.
	handleIndex := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Welcome to ebs-snapshot-exporter. Visit /metrics."))
	})
	// Initialize router and define all endpoints.
	router := http.NewServeMux()
	router.Handle("/", handleIndex)
	router.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	// Initialize server.
	server := &http.Server{
		Addr:         hub.config.Server.Address,
		Handler:      router,
		ReadTimeout:  hub.config.Server.ReadTimeout * time.Millisecond,
		WriteTimeout: hub.config.Server.WriteTimeout * time.Millisecond,
	}
	// Start the server. Blocks the main thread.
	hub.logger.Infof("starting server listening on %v", hub.config.Server.Address)
	if err := server.ListenAndServe(); err != nil {
		hub.logger.Fatalf("error starting server: %v", err)
	}
}
