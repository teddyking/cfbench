package bench

import (
	"fmt"
	"math"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
)

func ExtractBenchmarkScale(appGUID string, instances int) Phases {
	phases := Phases{
		&Phase{
			Name:             "ScaleTotal",
			StartMsg:         fmt.Sprintf(`Updated app with guid %s ({"instances"=>%d})`, appGUID, instances),
			EndMsg:           "Container became healthy",
			ShortName:        "total-scale",
			EndMsgOccurences: instances - 1,
		},
	}

	return phases
}

func ExtractBenchmarkPush(appGUID string, instances int) Phases {
	phases := Phases{
		&Phase{
			Name:      "Total",
			StartMsg:  "Created app with guid " + appGUID,
			EndMsg:    "Container became healthy",
			ShortName: "total",
		},
		&Phase{
			Name:       "Staging",
			StartMsg:   "Creating container",
			EndMsg:     "Successfully destroyed container",
			SourceType: "STG",
			ShortName:  "staging",
		},
		&Phase{
			Name:      "Upload droplet",
			StartMsg:  "Uploading droplet, build artifacts cache...",
			EndMsg:    "Uploading complete",
			ShortName: "upload-droplet",
		},
		&Phase{
			Name:       "Total run",
			StartMsg:   "Creating container",
			EndMsg:     "Container became healthy",
			SourceType: "CELL",
			ShortName:  "total-run",
		},
		&Phase{
			Name:       "Creating run container",
			StartMsg:   "Creating container",
			EndMsg:     "Successfully created container",
			SourceType: "CELL",
			ShortName:  "creating-run-container",
		},
		&Phase{
			Name:      "Health check",
			StartMsg:  "Starting health monitoring of container",
			EndMsg:    "Container became healthy",
			ShortName: "health-check",
		},
	}

	return phases
}

type Phase struct {
	Name             string
	StartMsg         string
	EndMsg           string
	SourceType       string
	ShortName        string
	EndMsgOccurences int

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

func (p Phases) PopulateTimestamps(appGUID string, events []*events.Envelope) {
	for _, phase := range p {
		remainingOccurences := math.Max(float64(phase.EndMsgOccurences), 1)
		for _, event := range events {
			logMsg := event.GetLogMessage()

			if logMsg.GetAppId() != appGUID {
				continue
			}

			if phase.SourceType != "" && phase.SourceType != logMsg.GetSourceType() {
				continue
			}

			logLine := logMsg.GetMessage()
			//fmt.Printf("%s/%v: [%s]\n", logMsg.GetSourceType(), logMsg.GetSourceInstance(), string(logLine))

			if phase.StartMsg == string(logLine) {
				phase.StartTimestamp = *logMsg.Timestamp
			} else if phase.EndMsg == string(logLine) {
				remainingOccurences--
				if remainingOccurences == 0 {
					phase.EndTimestamp = *logMsg.Timestamp
					break
				}
			}
		}
	}
}
