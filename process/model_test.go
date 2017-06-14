package process

import (
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("process verfication", func() {

	Context("given two logs messages of a step", func() {
		var (
			s                *step
			logStart, logEnd *events.LogMessage
		)
		BeforeEach(func() {
			s = &step{startMsg: "Step X starts", endMsg: "Step X ends"}
			logStart = createLogMsg("Step X starts", "whatever", 1000)
			logEnd = createLogMsg("Step X ends", "whatever", 1900)
		})
		JustBeforeEach(func() {
			s.matchMe(logStart)
			s.matchMe(logEnd)
		})
		It("matches the timestamp to start/end times", func() {
			Expect(s.startTime).To(Equal(int64(1000)))
			Expect(s.endTime).To(Equal(int64(1900)))
		})
	})

	Context("given a sequence of envelope messages", func() {
		var (
			s    string
			p    process
			msgs []*events.Envelope
		)

		BeforeEach(func() {
			s = "Staging"
			p = NewPushProcess("123456")
		})

		JustBeforeEach(func() {
			p.GetTimestamps(msgs)
		})
		Context("when the app guid is correct", func() {
			BeforeEach(func() {
				msgs = []*events.Envelope{
					createEnvelopeMsg("Staging...", "123456", 2000),
					createEnvelopeMsg("Staging complete", "123456", 2400),
				}
			})
			It("finds the correct duration of the step", func() {
				Expect(p.steps[s].getDuration()).To(Equal(int64(400)))
			})
		})

		Context("when the app guid is NOT correct", func() {
			BeforeEach(func() {
				msgs = []*events.Envelope{
					createEnvelopeMsg("Staging...", "98765", 2000),
					createEnvelopeMsg("Staging complete", "98765", 2400),
				}
			})
			It("does not find a duration of the step for that guid", func() {
				Expect(p.steps[s].getDuration()).NotTo(Equal(int64(400)))
			})
		})
	})
})

func createEnvelopeMsg(message, guid string, timestamp int64) *events.Envelope {
	return &events.Envelope{
		LogMessage: createLogMsg(message, guid, timestamp),
	}
}
func createLogMsg(message, guid string, timestamp int64) *events.LogMessage {
	return &events.LogMessage{
		Message:     []byte(message),
		MessageType: events.LogMessage_OUT.Enum(),
		AppId:       proto.String(guid),
		SourceType:  proto.String("DEA"),
		Timestamp:   proto.Int64(timestamp),
	}
}
