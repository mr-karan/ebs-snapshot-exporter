package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/mr-karan/ebs-exporter/internal/ebs"
	"github.com/mr-karan/ebs-exporter/internal/metrics"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

// initLogger initializes logger instance.
func initLogger() *logrus.Logger {
	logger := logrus.New()

	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:          true,
		DisableLevelTruncation: true,
	})

	return logger
}

// initConfig loads config to `ko` object.
func initConfig(lo *logrus.Logger, cfgDefault string, envPrefix string) (*koanf.Koanf, error) {
	var (
		ko = koanf.New(".")
		f  = flag.NewFlagSet("ebs_exporter", flag.ContinueOnError)
	)

	// Configure Flags.
	f.Usage = func() {
		fmt.Println(f.FlagUsages())
		os.Exit(0)
	}

	// Register `--config` flag.
	cfgPath := f.String("config", cfgDefault, "Path to a config file to load.")

	// Parse and Load Flags.
	err := f.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	// Load the config files from the path provided.
	lo.WithField("path", *cfgPath).Info("attempting to load config from file")

	err = ko.Load(file.Provider(*cfgPath), toml.Parser())
	if err != nil {
		// If the default config is not present, print a warning and continue reading the values from env.
		if *cfgPath == cfgDefault {
			lo.WithError(err).Warn("unable to open sample config file")
		} else {
			return nil, err
		}
	}

	lo.Info("attempting to read config from env vars")
	// Load environment variables if the key is given
	// and merge into the loaded config.
	if envPrefix != "" {
		err = ko.Load(env.Provider(envPrefix, ".", func(s string) string {
			return strings.Replace(strings.ToLower(
				strings.TrimPrefix(s, envPrefix)), "__", ".", -1)
		}), nil)
		if err != nil {
			return nil, err
		}
	}

	return ko, nil
}

// initEBS initializes a AWS manager.
func initEBS(ko *koanf.Koanf, lo *logrus.Logger) *ebs.Manager {
	ebs, err := ebs.New(ebs.Opts{
		Region:    ko.String("aws.region"),
		AccessKey: ko.String("aws.access_key"),
		SecretKey: ko.String("aws.secret_key"),
		RoleARN:   ko.String("aws.role_arn"),
		Logger:    lo,
	})

	if err != nil {
		lo.WithError(err).Fatal("error initialising notifier")
	}

	return ebs
}

// initMetrics initializes a Metrics manager.
func initMetrics() *metrics.Manager {
	return metrics.New("ebs")
}

func initOpts(ko *koanf.Koanf, lo *logrus.Logger) Opts {
	filters := make([]ebs.Filters, 0)
	err := ko.Unmarshal("aws.filters", &filters)
	if err != nil {
		lo.Fatal("error unmarshalling filters")
	}

	return Opts{
		CollectionInterval: ko.Duration("app.metrics_collect_interval"),
		Filters:            filters,
	}
}
