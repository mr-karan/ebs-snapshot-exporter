package main

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "ebs_snapshots"
)

// NewExporter returns an initialized `Exporter`.
func (hub *Hub) NewExporter(job *Job) (*Exporter, error) {
	ec2, err := hub.NewEC2Client(&job.AWSCreds)
	if err != nil {
		hub.logger.Errorf("Error initializing EC2 Client")
		return nil, err
	}
	return &Exporter{
		Mutex:  sync.Mutex{},
		client: ec2,
		job:    job,
		hub:    hub,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, job.Name, "up"),
			"Could the AWS EC2 API be reached.",
			nil,
			nil,
		),
		version: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, job.Name, "version"),
			"Version of ebs-snapshot-exporter",
			[]string{"build"},
			nil,
		),
		snapshotsCount:     constructPromMetric("count", job.Name, "The total number of snapshots", job.ExportedTags),
		snapshotVolumeSize: constructPromMetric("volume_size", job.Name, "Size of volume assosicated with the EBS snapshot", job.ExportedTags),
		snapshotStartTime:  constructPromMetric("start_time", job.Name, "Start Timestamp of EBS Snapshot", job.ExportedTags),
	}, nil
}

// sendSafeMetric is a concurrent safe method to send metrics to a channel. Since we are collecting metrics from AWS API, there might be possibility where
// a timeout occurs from Prometheus' collection context and the channel is closed but Goroutines running in background can still
// send metrics to this closed channel which would result in panic and crash. To solve that we use context and check if the channel is not closed
// and only send the metrics in that case. Else it logs the error and returns in a safe way.
func (hub *Hub) sendSafeMetric(ctx context.Context, ch chan<- prometheus.Metric, metric prometheus.Metric) error {
	// Check if collection context is finished
	select {
	case <-ctx.Done():
		// don't send metrics, instead return in a "safe" way
		hub.logger.Errorf("Attempted to send metrics to a closed channel after collection context had finished: %s", metric)
		return ctx.Err()
	default: // continue
	}
	// Send metrics if collection context is still open
	ch <- metric
	return nil
}

// Describe describes all the metrics ever exported by the exporter. It implements `prometheus.Collector`.
func (p *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- p.version
	ch <- p.up
	ch <- p.snapshotsCount
	ch <- p.snapshotStartTime
	ch <- p.snapshotVolumeSize
}

// Collect is called by the Prometheus registry when collecting
// metrics. This method may be called concurrently and must therefore be
// implemented in a concurrency safe way. It implements `prometheus.Collector`
func (p *Exporter) Collect(ch chan<- prometheus.Metric) {
	// Initialize context to keep track of the collection.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Lock the exporter for one iteration of collection as `Collect` can be called concurrently.
	p.Lock()
	defer p.Unlock()

	// Fetch snapshots data from EC2 API.
	snaps, err := p.client.GetSnapshots(p.job.Filters)
	if err != nil {
		p.hub.logger.Errorf("Error collecting metrics from EC2 API: %v", err)
		p.hub.sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(p.up, prometheus.GaugeValue, 0))
		return
	}
	p.collectSnapshotMetrics(ctx, ch, snaps)
	// Send default metrics data.
	p.hub.sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(p.version, prometheus.GaugeValue, 1, p.hub.version))
	p.hub.sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(p.up, prometheus.GaugeValue, 1))
}

func (p *Exporter) collectSnapshotMetrics(ctx context.Context, ch chan<- prometheus.Metric, snaps *ec2.DescribeSnapshotsOutput) {
	// Iterate through all snapshots and collect only those which match the user defined tags.
	for _, s := range snaps.Snapshots {
		// Initialize common label values for all metrics exported below.
		exportedLabelValues := []string{*s.SnapshotId, *s.VolumeId, p.job.AWSCreds.Region, *s.Progress, *s.State}
		// Initialize an empty slice for all additional label names.
		exportedLabelNames := []string{}
		// Iterate through a set of tags and append to the slice of exportedLabelValues if the tag key matches.
		// with user defined tags
		for _, t := range s.Tags {
			for _, k := range p.job.ExportedTags {
				if *t.Key == k {
					// If the tag matches, add the value to exportedLabelValues.
					exportedLabelValues = append(exportedLabelValues, *t.Value)
					// Also add the tag key to exported label names of the metric.
					exportedLabelNames = append(exportedLabelNames, k)
				}
			}
		}
		// Create a map to indicate if a label name is present.
		set := make(map[string]bool)
		for _, v := range exportedLabelNames {
			set[v] = true
		}
		// For all the other user defined tags if there is no value present send empty string to maintain
		// label cardinality.
		for _, k := range p.job.ExportedTags {
			if !set[k] {
				exportedLabelValues = append(exportedLabelValues, "")
				exportedLabelNames = append(exportedLabelNames, k)
			}
		}
		// At this point exportedLabelNames should have all label names for the metric and exportedLabelValues should have consistent data.
		snapVol := constructPromMetric("volume_size", p.job.Name, "Size of volume assosicated with the EBS snapshot", exportedLabelNames)
		p.hub.sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(snapVol, prometheus.GaugeValue, float64(*s.VolumeSize), exportedLabelValues...))
		snapStartTime := constructPromMetric("start_time", p.job.Name, "Start Timestamp of EBS Snapshot", exportedLabelNames)
		p.hub.sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(snapStartTime, prometheus.GaugeValue, float64(s.StartTime.Unix()), exportedLabelValues...))
	}
}

// Returns an intialized prometheus.Desc instance
func constructPromMetric(metricName string, jobName string, helpText string, additionalLabels []string) *prometheus.Desc {
	// Default labels for any metric constructed with this function.
	labels := []string{"snapshot_id", "vol_id", "region", "progress", "state"}
	// Iterate through a slice of additional labels to be exported.
	for _, k := range additionalLabels {
		// Replace all tags with underscores if present to make it a valid Prometheus label name.
		labels = append(labels, replaceWithUnderscores(k))
	}
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", metricName),
		helpText,
		labels, prometheus.Labels{"job": jobName},
	)
}
