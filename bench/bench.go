package bench

import (
	"fmt"
	"os"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
)

func ExtractBenchmark(appGUID string, events []*events.Envelope, instances int) Phases {
	phases := Phases{
		&Phase{
			Name:      "Total",
			startMsg:  "Created app with guid " + appGUID,
			endMsg:    "Container became healthy",
			ShortName: "total",
			waitFor:   1,
		},
		&Phase{
			Name:       "Staging",
			startMsg:   "Creating container",
			endMsg:     "Successfully destroyed container",
			sourceType: "STG",
			ShortName:  "staging",
		},
		&Phase{
			Name:      "Upload droplet",
			startMsg:  "Uploading droplet, build artifacts cache...",
			endMsg:    "Uploading complete",
			ShortName: "upload-droplet",
		},
		&Phase{
			Name:       "Scaling",
			startMsg:   fmt.Sprintf(`Updated app with guid %s ({"instances"=>%d})`, appGUID, instances),
			endMsg:     "----",
			sourceType: "CELL",
			ShortName:  "scale",
			waitFor:    instances - 1,
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
	waitFor    int

	StartTimestamp int64
	EndTimestamp   int64
}

func (p Phase) IsValid() bool {
	return p.EndTimestamp != 0 && p.StartTimestamp != 0
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
			fmt.Fprintf(os.Stderr, "--> MESSAGE: %s\n", logLine)

			if phase.startMsg == logLine {
				phase.StartTimestamp = *logMsg.Timestamp
			} else if phase.endMsg == logLine {
				phase.EndTimestamp = *logMsg.Timestamp
				break
			}
		}
	}
}
