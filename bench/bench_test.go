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
			createEnvelopeMsg("Creating container", "123456", "STG", 2000),
			createEnvelopeMsg("Successfully destroyed container", "123456", "STG", 2400),
		}
	})

	JustBeforeEach(func() {
		phases = bench.ExtractBenchmark(appGUID, msgs)
	})

	Context("when the app guid matches a phase", func() {
		BeforeEach(func() {
			appGUID = "123456"
		})

		It("finds the correct duration of the phase", func() {
			Expect(phases[1].Name).To(Equal("Staging"))
			Expect(phases[1].Duration()).To(Equal(time.Duration(400)))
		})
	})

	Context("when the app guid doesn't match a phase", func() {
		BeforeEach(func() {
			appGUID = "garbage"
		})

		It("doesn't populate the duration", func() {
			Expect(phases[1].Name).To(Equal("Staging"))
			Expect(phases[1].Duration()).To(Equal(time.Duration(0)))
		})
	})

	Context("when the source type matches", func() {
		BeforeEach(func() {
			appGUID = "123456"
			msgs = []*events.Envelope{
				createEnvelopeMsg("Creating container", "123456", "CELL", 2000),
				createEnvelopeMsg("Successfully created container", "123456", "CELL", 2001),
			}
		})

		It("finds the correct duration of the phase", func() {
			Expect(phases[4].Name).To(Equal("Creating run container"))
			Expect(phases[4].Duration()).To(Equal(time.Duration(1)))
		})
	})

	Context("when the source type doesn't match", func() {
		BeforeEach(func() {
			appGUID = "123456"
			msgs = []*events.Envelope{
				createEnvelopeMsg("Creating container", "123456", "", 2000),
				createEnvelopeMsg("Successfully created container", "123456", "", 2001),
			}
		})

		It("doesn't populate the duration", func() {
			Expect(phases[4].Name).To(Equal("Creating run container"))
			Expect(phases[4].Duration()).To(Equal(time.Duration(0)))
		})
	})
})

func createEnvelopeMsg(message, guid, sourceType string, timestamp int64) *events.Envelope {
	return &events.Envelope{
		LogMessage: &events.LogMessage{
			Message:     []byte(message),
			MessageType: events.LogMessage_OUT.Enum(),
			AppId:       proto.String(guid),
			SourceType:  proto.String(sourceType),
			Timestamp:   proto.Int64(timestamp),
		},
	}
}
