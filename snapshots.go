package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// FetchAWSData represents the set of methods used to interact with AWS API.
type FetchAWSData interface {
	GetSnapshots(awsFilters []Filters) (*ec2.DescribeSnapshotsOutput, error)
}

// NewEC2Client creates an AWS Session and returns an initialized client with the session object embedded.
func (hub *Hub) NewEC2Client(awsCreds *AWSCreds) (*EC2Client, error) {
	// Initialize default config.
	config := &aws.Config{
		Region: aws.String(awsCreds.Region),
	}
	// Override Access Key and Secret Key env vars if specified in config.
	if awsCreds.AccessKey != "" && awsCreds.SecretKey != "" {
		config.Credentials = credentials.NewStaticCredentials(awsCreds.AccessKey, awsCreds.SecretKey, "")
	}
	// Initialize session with custom config embedded.
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: *config,
	})
	if err != nil {
		hub.logger.Errorf("Error creating AWS Session %s", err)
		return nil, fmt.Errorf("could not create aws session")
	}
	// Initialize EC2 Client.
	var ec2Client *ec2.EC2
	if awsCreds.RoleARN != "" {
		// Assume Role if specified
		hub.logger.Debugf("Assuming Role: %v", awsCreds.RoleARN)
		creds := stscreds.NewCredentials(sess, awsCreds.RoleARN)
		ec2Client = ec2.New(sess, &aws.Config{Credentials: creds})
	} else {
		ec2Client = ec2.New(sess)
	}
	return &EC2Client{
		client: ec2Client,
	}, nil
}

// GetSnapshots takes awsFilters as input and returns the API response of `DescribeSnapshots` API Call.
func (e *EC2Client) GetSnapshots(awsFilters []Filters) (*ec2.DescribeSnapshotsOutput, error) {
	filters := make([]*ec2.Filter, 0, len(awsFilters))
	for _, tag := range awsFilters {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String(tag.Name),
			Values: []*string{aws.String(tag.Value)},
		})
	}
	// Construct request params for the API Request.
	params := &ec2.DescribeSnapshotsInput{}
	if len(filters) != 0 {
		params = &ec2.DescribeSnapshotsInput{Filters: filters}
	}
	resp, err := e.client.DescribeSnapshots(params)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
