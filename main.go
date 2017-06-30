package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/teddyking/cfbench/bench"
	"github.com/teddyking/cfbench/cf"
)

func main() {
	authToken := envMustHave("CF_AUTH_TOKEN")

	pwd, err := os.Getwd()
	mustNot("get CWD", err)
	appDir := flag.String("app-dir", pwd, "The directory of the app to push")
	dopplerAddress := flag.String("doppler-address", "", "doppler address")

	flag.Parse()

	if *dopplerAddress == "" {
		fmt.Println("must set --doppler-address")
		os.Exit(1)
	}

	fmt.Println("Buffering all messages from Firehose in the background.")
	firehoseEvents := make([]*events.Envelope, 100)
	cnsmr := consumer.New(*dopplerAddress, &tls.Config{InsecureSkipVerify: true}, nil)
	defer cnsmr.Close()
	stopFirehose := make(chan struct{})
	go func(stop <-chan struct{}) {
		msgChan, errChan := cnsmr.Firehose("cfbench", string(authToken))
		for {
			select {
			case msg := <-msgChan:
				firehoseEvents = append(firehoseEvents, msg)
			case err := <-errChan:
				mustNot("consuming firehose", err)
			case <-stop:
				return
			}
		}
	}(stopFirehose)

	appName := "benchme"
	must("pushing app", cf.Push(appName, *appDir))
	appGuid, err := cf.AppGuid(appName)
	mustNot("getting app GUID", err)
	must("deleting app", cf.Delete(appName))

	fmt.Println("Waiting a few seconds in case some relevant messages are late")
	time.Sleep(time.Second * 5)

	close(stopFirehose)

	fmt.Printf("\nResults:\n")

	for _, phase := range bench.ExtractBenchmark(appGuid, firehoseEvents) {
		fmt.Printf("%s: %s (%s - %s)\n", phase.Name, phase.Duration().String(),
			time.Unix(0, phase.StartTimestamp), time.Unix(0, phase.EndTimestamp))
	}
}

func envMustHave(key string) string {
	value := os.Getenv(key)
	if value == "" {
		fmt.Printf("please set %s\n", key)
		os.Exit(1)
	}
	return value
}

func mustNot(action string, err error) {
	if err != nil {
		fmt.Printf("error %s: %s\n", action, err)
		os.Exit(1)
	}
}

var must = mustNot
