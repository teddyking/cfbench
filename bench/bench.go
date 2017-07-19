package bench

import (
	"time"

	"github.com/cloudfoundry/sonde-go/events"
)

func ExtractBenchmark(appGUID string, events []*events.Envelope) Phases {
	phases := Phases{
		&Phase{
			Name:      "Total",
			startMsg:  "Created app with guid " + appGUID,
			endMsg:    "Container became healthy",
			ShortName: "total",
		},
		&Phase{
			Name:      "Staging",
			startMsg:  "Staging...",
			endMsg:    "Staging complete",
			ShortName: "staging",
		},
		&Phase{
			Name:      "Upload droplet",
			startMsg:  "Uploading droplet, build artifacts cache...",
			endMsg:    "Uploading complete",
			ShortName: "upload-droplet",
		},
		&Phase{
			Name:       "Total run",
			startMsg:   "Creating container",
			endMsg:     "Container became healthy",
			sourceType: "CELL",
			ShortName:  "total-run",
		},
		&Phase{
			Name:       "Creating run container",
			startMsg:   "Creating container",
			endMsg:     "Successfully created container",
			sourceType: "CELL",
			ShortName:  "creating-run-container",
		},
		&Phase{
			Name:      "Health check",
			startMsg:  "Starting health monitoring of container",
			endMsg:    "Container became healthy",
			ShortName: "health-check",
		},
	}

	phases.populateTimestamps(appGUID, events)
	return phases
}

type Phase struct {
	Name       string
	startMsg   string
	endMsg     string
	sourceType string
	ShortName  string

	StartTimestamp int64
	EndTimestamp   int64
}

func (p Phase) Duration() time.Duration {
	return time.Duration(p.EndTimestamp - p.StartTimestamp)
}

type Phases []*Phase

func (p Phases) populateTimestamps(appGUID string, events []*events.Envelope) {
	for _, phase := range p {
		for _, event := range events {
			logMsg := event.GetLogMessage()

			if logMsg.GetAppId() != appGUID {
				continue
			}

			if phase.sourceType != "" && phase.sourceType != *logMsg.SourceType {
				continue
			}

			logLine := string(logMsg.Message)
			if phase.startMsg == logLine {
				phase.StartTimestamp = *logMsg.Timestamp
			} else if phase.endMsg == logLine {
				phase.EndTimestamp = *logMsg.Timestamp
			}
		}
	}
}
