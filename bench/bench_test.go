package bench_test

import (
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/teddyking/cfbench/bench"
)

var _ = Describe("ExtractBenchmark", func() {
	var (
		appGUID string
		msgs    []*events.Envelope

		phases bench.Phases
	)

	BeforeEach(func() {
		msgs = []*events.Envelope{
			createEnvelopeMsg("Staging...", "123456", 2000),
			createEnvelopeMsg("Staging complete", "123456", 2400),
		}
	})

	JustBeforeEach(func() {
		phases = bench.ExtractBenchmark(appGUID, msgs)
	})

	Context("when the app guid matches a phase", func() {
		BeforeEach(func() {
			appGUID = "123456"
		})

		It("finds the correct duration of the step", func() {
			Expect(phases).To(HaveLen(3))
			Expect(phases[1].Name).To(Equal("Staging"))
			Expect(phases[1].Duration()).To(Equal(time.Duration(400)))
		})
	})

	Context("when the app guid doesn't match a phase", func() {
		BeforeEach(func() {
			appGUID = "garbage"
		})

		It("does not find a duration of the step for that guid", func() {
			Expect(phases).To(HaveLen(3))
			Expect(phases[1].Name).To(Equal("Staging"))
			Expect(phases[1].Duration()).NotTo(Equal(time.Duration(400)))
		})
	})
})

func createEnvelopeMsg(message, guid string, timestamp int64) *events.Envelope {
	return &events.Envelope{
		LogMessage: &events.LogMessage{
			Message:     []byte(message),
			MessageType: events.LogMessage_OUT.Enum(),
			AppId:       proto.String(guid),
			SourceType:  proto.String("DEA"),
			Timestamp:   proto.Int64(timestamp),
		},
	}
}
