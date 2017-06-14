package process

import (
	"fmt"
	"time"

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

func (s step) getDuration() time.Duration {
	return time.Duration(s.endTime - s.startTime)
}

func (s step) getProzent(total time.Duration) float64 {
	return (float64(s.endTime) - float64(s.startTime)) * 100 / float64(total)
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
	fmt.Printf("\n###################################################\n")
	fmt.Printf("SUMMARY\n")
	total := p.steps["Total"].getDuration() //FIXME. this is not safe, Total key might not exist
	fmt.Printf("%-20s: %15v \n", "Step", "Duration (%)")
	fmt.Printf("%-20s: %15v\n", "---", "---")
	for k, v := range p.steps {
		fmt.Printf("%-20s: %5.2f sec (%5.2f%%)\n", k, v.getDuration().Seconds(), v.getProzent(total))
	}
	fmt.Printf("\n###################################################\n")
}

func InvestigateMessages(envelopes []*events.Envelope, appGuid string) {
	//reader := bufio.NewReader(os.Stdin)
	fmt.Println("Reading messages")
	fmt.Printf("%+15s %+15s %+15s %+15s %+15s %+15s\n", "Timestamp", "eventType", "origin", "job", "index", "ip")
	for _, e := range envelopes {
		fmt.Printf("%+15v %+15v %+15v %+15v %+15v %+15v\t",
			e.GetTimestamp(), e.GetEventType(), e.GetOrigin(), e.GetJob(), e.GetIndex(), e.GetIp())
		switch e.GetEventType() {
		case events.Envelope_LogMessage:
			fmt.Printf("%v\n", e.GetLogMessage())
		case events.Envelope_ValueMetric:
			fmt.Printf("%v\n", e.GetValueMetric())
		case events.Envelope_CounterEvent:
			fmt.Printf("%v\n", e.GetCounterEvent())
		default:
			fmt.Printf("\n")
		}
		//reader.ReadString('\n')
	}
}
