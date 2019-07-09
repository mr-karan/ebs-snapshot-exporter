package main

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Hub represents the structure for all app wide functions and structs
type Hub struct {
	logger  *logrus.Logger
	config  config
	version string
}

// cfgApp represents the structure to hold App specific configuration.
type cfgApp struct {
	LogLevel string `koanf:"log_level"`
	Jobs     []Job  `koanf:"jobs"`
}

// cfgServer represents the structure to hold Server specific configuration
type cfgServer struct {
	Name         string        `koanf:"name"`
	Address      string        `koanf:"address"`
	ReadTimeout  time.Duration `koanf:"read_timeout"`
	WriteTimeout time.Duration `koanf:"write_timeout"`
	MaxBodySize  int           `koanf:"max_body_size"`
}

// config represents the structure to hold configuration loaded from an external data source.
type config struct {
	App    cfgApp    `koanf:"app"`
	Server cfgServer `koanf:"server"`
}

// Filters represents the structure to hold the AWS Filters Data required to filter snapshots.
// Filter API documentation: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_Filter.html.
type Filters struct {
	Name  string `koanf:"name"`
	Value string `koanf:"value"`
}

// AWSCreds represents the structure to hold AWS Credentials required to create AWS session.
type AWSCreds struct {
	Region    string `koanf:"region"`
	RoleARN   string `koanf:"role_arn"`
	AccessKey string `koanf:"access_key"`
	SecretKey string `koanf:"secret_key"`
}

// Job represents a list of arbitary key value pair used to filter EBS Snapshots.
type Job struct {
	Name         string    `koanf:"name"`
	AWSCreds     AWSCreds  `koanf:"aws_creds"`
	Filters      []Filters `koanf:"filters"`
	ExportedTags []string  `koanf:"exported_tags"`
}

// Exporter represents the structure to hold Prometheus Descriptors. It implements prometheus.Collector
type Exporter struct {
	sync.Mutex                          // Lock exporter to protect from concurrent scrapes.
	hub                *Hub             // To access logger and other app wide config.
	client             FetchAWSData     // Implements FetchAWSData interface which is a set of methods to interact with AWS.
	job                *Job             // Holds the Job metadata.
	up                 *prometheus.Desc // Represents if a scrape was successful or not.
	version            *prometheus.Desc // Represents verion of the exporter.
	snapshotsCount     *prometheus.Desc // Represents the total count of EBS Snapshots.
	snapshotVolumeSize *prometheus.Desc // Represents the size of the EBS Volume related to a particular snapshot.
	snapshotStartTime  *prometheus.Desc // Represents the start time of creating snapshots as recorded AWS.
}

// EC2Client represents the structure to hold EC2 Client object required to create AWS session and
// interact with AWS API SDK.
type EC2Client struct {
	client ec2iface.EC2API
}
