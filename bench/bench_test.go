package bench_test

import (
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/teddyking/cfbench/bench"
)

var _ = Describe("Phase", func() {
	var (
		appGUID string
		phases  Phases
		msgs    []*events.Envelope
	)

	JustBeforeEach(func() {

	})

	Context("filters by app guid", func() {
		BeforeEach(func() {
			phases = Phases{
				&Phase{
					Name:             "Total",
					StartMsg:         "start-message",
					EndMsg:           "end-message",
					ShortName:        "total",
					SourceType:       "FOO",
					EndMsgOccurences: 1,
				},
			}

			msgs = []*events.Envelope{
				createEnvelopeMsg("start-message", "123456", "FOO", 2000),
				createEnvelopeMsg("end-message", "123456", "FOO", 2400),
			}
		})

		It("finds the phases which match", func() {
			phases.PopulateTimestamps("123456", msgs)
			Expect(phases[0].StartTimestamp).To(Equal(int64(2000)))
			Expect(phases[0].EndTimestamp).To(Equal(int64(2400)))
		})

		It("ignores the phases which don't match", func() {
			phases.PopulateTimestamps("garbage", msgs)
			Expect(phases[0].StartTimestamp).To(Equal(int64(0)))
			Expect(phases[0].EndTimestamp).To(Equal(int64(0)))
		})

	})

	Context("filters by source type", func() {
		BeforeEach(func() {
			phases = Phases{
				&Phase{
					Name:             "Total",
					StartMsg:         "puppy",
					EndMsg:           "kitten",
					ShortName:        "total",
					SourceType:       "FOO",
					EndMsgOccurences: 1,
				},
				&Phase{
					Name:             "Total",
					StartMsg:         "puppy",
					EndMsg:           "kitten",
					ShortName:        "total",
					SourceType:       "BAR",
					EndMsgOccurences: 1,
				},
			}

			msgs = []*events.Envelope{
				createEnvelopeMsg("puppy", "123456", "FOO", 2000),
				createEnvelopeMsg("kitten", "123456", "FOO", 2001),
			}
			appGUID = "123456"
		})

		It("finds the phases which match", func() {
			phases.PopulateTimestamps(appGUID, msgs)
			Expect(phases[0].StartTimestamp).To(Equal(int64(2000)))
			Expect(phases[0].EndTimestamp).To(Equal(int64(2001)))
		})
		It("ignores the phases which don't match", func() {
			phases.PopulateTimestamps(appGUID, msgs)
			Expect(phases[1].StartTimestamp).To(Equal(int64(0)))
			Expect(phases[1].EndTimestamp).To(Equal(int64(0)))
		})
	})

	Context("when endMsgOccurences is set", func() {
		BeforeEach(func() {
			phases = Phases{
				&Phase{
					Name:             "Total",
					StartMsg:         "puppy",
					EndMsg:           "kitten",
					ShortName:        "total",
					SourceType:       "FOO",
					EndMsgOccurences: 10,
				},
			}

			for i := 0; i <= 10; i++ {
				msgs = append(msgs, []*events.Envelope{
					createEnvelopeMsg("puppy", "123456", "FOO", 2000),
					createEnvelopeMsg("kitten", "123456", "FOO", 2001+int64(i)),
				}...)
			}
			appGUID = "123456"
		})

		It("sets the EndTimestamp to the occurrence of the last end message", func() {
			phases.PopulateTimestamps(appGUID, msgs)
			Expect(phases[0].StartTimestamp).To(Equal(int64(2000)))
			Expect(phases[0].EndTimestamp).To(Equal(int64(2009)))
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
