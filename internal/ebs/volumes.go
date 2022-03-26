package ebs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func (m *Manager) FetchVolumes(ebsFilters []Filters) (*ec2.DescribeVolumesOutput, error) {
	filters := make([]types.Filter, 0, len(ebsFilters))

	for _, tag := range ebsFilters {
		filters = append(filters, types.Filter{
			Name:   aws.String(tag.Name),
			Values: []string{tag.Value},
		})
	}

	// Construct request params for the API Request.
	params := &ec2.DescribeVolumesInput{}
	if len(filters) != 0 {
		params = &ec2.DescribeVolumesInput{Filters: filters}
	}

	return m.ec2.DescribeVolumes(context.TODO(), params)
}
