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
		"Staging": &step{startMsg: "Staging...", endMsg: "Staging complete"},
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
	fmt.Printf("SUMMARY\n")
	fmt.Printf("App GUID = %s\n", p.appGuid)
	fmt.Printf("[ %s ]: [ %v ]\n", "Step", "Duration")
	for k, v := range p.steps {
		fmt.Printf("[%s]: [ %v]\n", k, v.getDuration())
	}
}

// func processMessages(messages []*events.Envelope, appGuid string) {
// 	reader := bufio.NewReader(os.Stdin)
// 	fmt.Println("Reading messages")
// 	for _, m := range messages {
// 		if m.GetLogMessage().GetAppId() == appGuid {
// 			fmt.Println(m.String())
// 			reader.ReadString('\n')
// 		} else {
// 			fmt.Printf("[%s]==[%s]\n", m.GetLogMessage().GetAppId(), appGuid)
// 		}
// 	}
// }
