package ebs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/sirupsen/logrus"
)

type Opts struct {
	Region    string
	AccessKey string
	SecretKey string
	RoleARN   string
	Logger    *logrus.Logger
}

type Manager struct {
	lo  *logrus.Logger
	ec2 *ec2.Client
	cw  *cloudwatch.Client
}

// Filters represents the structure to hold the AWS Filters Data required to filter snapshots.
// Filter API documentation: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_Filter.html.
type Filters struct {
	Name  string `koanf:"name"`
	Value string `koanf:"value"`
}

func New(opts Opts) (*Manager, error) {
	// Initialise client for interacting with EC2 APIs.

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		// handle error
	}
	ec2 := ec2.NewFromConfig(cfg)
	cw := cloudwatch.NewFromConfig(cfg)

	return &Manager{
		ec2: ec2,
		cw:  cw,
		lo:  opts.Logger,
	}, nil
}
