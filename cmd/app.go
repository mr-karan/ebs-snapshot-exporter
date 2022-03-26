package main

import (
	"fmt"
	"time"

	"github.com/mr-karan/ebs-exporter/internal/ebs"
	"github.com/mr-karan/ebs-exporter/internal/metrics"
	"github.com/sirupsen/logrus"
)

type Opts struct {
	CollectionInterval time.Duration
	Filters            []ebs.Filters
}

// App is the global contains
// instances of various objects used in the lifecyle of program.
type App struct {
	lo      *logrus.Logger
	metrics *metrics.Manager
	ebs     *ebs.Manager
	opts    Opts
}

var (
	labelAllocatedIOPS = `iops_allocated{id="%s", name="%s", az="%s"}`
	labelReadIOPS      = `iops_read{id="%s", name="%s", az="%s"}`
	labelWriteIOPS     = `iops_write{id="%s", name="%s", az="%s"}`
)

func (app *App) Collect() {
	var (
		collectTicker = time.NewTicker(app.opts.CollectionInterval).C
	)

	for range collectTicker {
		// Get the volumes.
		app.lo.Debug("fetching volumes")
		volumes, err := app.ebs.FetchVolumes(app.opts.Filters)
		if err != nil {
			app.lo.WithError(err).Error("error fetching volumes")
			continue
		}
		app.lo.WithField("count", len(volumes.Volumes)).Debug("fetched volumes")

		// Fetch "per volume" metric for each volume.
		for _, v := range volumes.Volumes {
			vol := v
			name := ""
			for _, t := range vol.Tags {
				if *t.Key == "Name" {
					name = *t.Value
				}
			}

			readIOPS, err := app.ebs.FetchIOPS("VolumeReadOps", *vol.VolumeId)
			if err != nil {
				app.lo.WithError(err).Error("error fetching VolumeReadOps")
				continue
			}
			writeIOPS, err := app.ebs.FetchIOPS("VolumeWriteOps", *vol.VolumeId)
			if err != nil {
				app.lo.WithError(err).Error("error fetching VolumeWriteOps")
				continue
			}

			// Publish metrics.
			app.metrics.Set(fmt.Sprintf(labelAllocatedIOPS, *vol.VolumeId, name, *vol.AvailabilityZone), float64(*vol.Iops))
			app.metrics.Set(fmt.Sprintf(labelReadIOPS, *vol.VolumeId, name, *vol.AvailabilityZone), readIOPS)
			app.metrics.Set(fmt.Sprintf(labelWriteIOPS, *vol.VolumeId, name, *vol.AvailabilityZone), writeIOPS)
		}
	}

}
