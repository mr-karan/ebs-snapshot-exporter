package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/sirupsen/logrus"
)

var (
	// Version of the build. This is injected at build-time.
	buildString = "unknown"
)

func main() {
	// Initialise logger.
	lo := initLogger()

	// Initialise and load the config.
	ko, err := initConfig(lo, "config.sample.toml", "EBS_EXPORTER_")
	if err != nil {
		// Need to `panic` since logger can only be initialised once config is initialised.
		panic(err.Error())
	}

	var (
		metrics = initMetrics()
		opts    = initOpts(ko, lo)
		ebs     = initEBS(ko, lo)
	)

	// Enable debug mode if specified.
	if ko.String("app.log") == "debug" {
		lo.SetLevel(logrus.DebugLevel)
	}

	app := &App{
		lo:      lo,
		ebs:     ebs,
		metrics: metrics,
		opts:    opts,
	}

	go app.Collect()

	app.lo.WithField("version", buildString).Info("booting ebs-exporter")

	// Initialise HTTP Router.
	r := chi.NewRouter()

	// Add some middlewares.
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	// Register Handlers
	r.Get("/", wrap(app, handleIndex))
	r.Get("/ping", wrap(app, handleHealthCheck))
	r.Get("/metrics", wrap(app, handleMetrics))

	// Start HTTP Server.
	app.lo.WithField("addr", ko.MustString("app.address")).Info("starting http server")
	srv := &http.Server{
		Addr:         ko.MustString("app.address"),
		ReadTimeout:  ko.MustDuration("app.server_timeout"),
		WriteTimeout: ko.MustDuration("app.server_timeout"),
		Handler:      r,
	}
	if err := srv.ListenAndServe(); err != nil {
		app.lo.WithError(err).Fatal("couldn't start server")
	}
}
