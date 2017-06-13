package process

import (
	"fmt"

	"github.com/cloudfoundry/sonde-go/events"
)

type step struct {
	startMsg, endMsg   string
	startTime, endTime int64
}

func (s *step) matchMe(m *events.LogMessage) {
	switch string(m.GetMessage()) {
	case s.startMsg:
		s.startTime = *m.Timestamp
	case s.endMsg:
		s.endTime = *m.Timestamp
	}
}

func (s step) getDuration() int64 {
	return s.endTime - s.startTime
}

type process struct {
	steps   map[string]*step
	appGuid string
}

func NewPushProcess(guid string) process {
	t := map[string]*step{
		"Total": &step{startMsg: "Created app with guid " + guid,
			endMsg: "Container became healthy"},
		"Staging": &step{startMsg: "Staging...",
			endMsg: "Staging complete"},
		"Upload Droplet": &step{startMsg: "Uploading droplet, build artifacts cache...",
			endMsg: "Uploading complete"},
	}

	return process{
		steps:   t,
		appGuid: guid,
	}
}

func (p *process) GetTimestamps(envelopes []*events.Envelope) {
	for _, step := range p.steps {
		for _, e := range envelopes {
			m := e.GetLogMessage()

			if m.GetAppId() != p.appGuid {
				continue
			}

			step.matchMe(m)
		}
	}
}

func (p process) PrintResult() {
	fmt.Printf("\n###############################\n")
	fmt.Printf("SUMMARY\n")
	fmt.Printf("[ %s ]: [ %v ]\n", "Step", "Duration")
	for k, v := range p.steps {
		fmt.Printf("[%s]: [ %v]\n", k, v.getDuration())
	}
	fmt.Printf("###############################\n")
}

func InvestigateMessages(envelopes []*events.Envelope, appGuid string) {
	//	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Reading messages")
	for _, e := range envelopes {
		m := e.GetLogMessage()
		if m.GetAppId() != appGuid {
			continue
		}
		fmt.Println(string(m.GetMessage()))
		//reader.ReadString('\n')
	}
}
