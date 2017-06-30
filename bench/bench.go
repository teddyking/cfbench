package bench

import (
	"time"

	"github.com/cloudfoundry/sonde-go/events"
)

func ExtractBenchmark(appGUID string, events []*events.Envelope) Phases {
	phases := Phases{
		&Phase{
			Name:     "Total",
			startMsg: "Created app with guid " + appGUID,
			endMsg:   "Container became healthy",
		},
		&Phase{
			Name:     "Staging",
			startMsg: "Staging...",
			endMsg:   "Staging complete",
		},
		&Phase{
			Name:     "Upload droplet",
			startMsg: "Uploading droplet, build artifacts cache...",
			endMsg:   "Uploading complete",
		},
	}

	phases.populateTimestamps(appGUID, events)
	return phases
}

type Phase struct {
	Name     string
	startMsg string
	endMsg   string

	startTimestamp int64
	endTimestamp   int64
}

func (p Phase) Duration() time.Duration {
	return time.Duration(p.endTimestamp - p.startTimestamp)
}

type Phases []*Phase

func (p Phases) populateTimestamps(appGUID string, events []*events.Envelope) {
	for _, phase := range p {
		for _, event := range events {
			logMsg := event.GetLogMessage()

			if logMsg.GetAppId() != appGUID {
				continue
			}

			logLine := string(logMsg.Message)
			if phase.startMsg == logLine {
				phase.startTimestamp = *logMsg.Timestamp
			} else if phase.endMsg == logLine {
				phase.endTimestamp = *logMsg.Timestamp
			}
		}
	}
}
