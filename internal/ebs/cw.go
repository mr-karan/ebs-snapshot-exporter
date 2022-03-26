package ebs

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

func (m *Manager) FetchIOPS(metric string, volumeID string) (float64, error) {
	input := &cloudwatch.GetMetricDataInput{
		EndTime:   aws.Time(time.Unix(time.Now().Unix(), 0)),
		StartTime: aws.Time(time.Unix(time.Now().Add(time.Duration(-5)*time.Minute).Unix(), 0)),
		ScanBy:    types.ScanBy(*aws.String("TimestampDescending")),
		MetricDataQueries: []types.MetricDataQuery{
			types.MetricDataQuery{
				Id: aws.String("getreadiops"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/EBS"),
						MetricName: aws.String(metric),
						Dimensions: []types.Dimension{
							types.Dimension{
								Name:  aws.String("VolumeId"),
								Value: aws.String(volumeID),
							},
						},
					},
					Period: aws.Int32(int32(300)),
					Stat:   aws.String("Sum"),
				},
			},
		},
	}

	result, err := m.cw.GetMetricData(context.TODO(), input)
	if err != nil {
		return 0, fmt.Errorf("Could not fetch metric data: %s", err)
	}

	// Volumes which are attached to instances won't have Monitoring data.
	if len(result.MetricDataResults) == 0 {
		return 0, nil
	}
	if len(result.MetricDataResults[0].Values) == 0 {
		return 0, nil
	}

	// Else, return the last value. Since we are always sorting by latest timestamp, we just
	// need to find the first element in metric and divide it by 300 to get a "per-second" count.
	iops := result.MetricDataResults[0].Values[0] / 300
	return iops, nil
}
