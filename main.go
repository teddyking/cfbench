package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/teddyking/cfbench/bench"
	"github.com/teddyking/cfbench/cf"
	"github.com/teddyking/cfbench/datadog"
)

func main() {
	appName := fmt.Sprintf("benchme-%v", time.Now().UnixNano())
	var appGuid string

	//Flag Part
	authToken := envMustHave("CF_AUTH_TOKEN")
	pwd, err := os.Getwd()
	mustNot("get CWD", err)
	action := flag.String("action", "push", "Push or scale")
	stack := flag.String("stack", "cflinuxfs2", "The stack to push the app to")
	buildpack := flag.String("buildpack", "", "The buildpack to push the app with")
	startCommand := flag.String("startCommand", "", "The start command to push the app with")
	appDir := flag.String("app-dir", pwd, "The directory of the app to push")
	dopplerAddress := flag.String("doppler-address", "", "doppler address")
	var instances int
	flag.IntVar(&instances, "instances", 1, "scale app after pushing")
	var tags tagList
	flag.Var(&tags, "tag", "a tag, can be specified multiple times")

	var jsonOutput bool
	flag.BoolVar(&jsonOutput, "json", false, "Generate datadog-compatible JSON output on stdout")

	flag.Parse()

	//Validation Part
	if *dopplerAddress == "" {
		log.Println("must set --doppler-address")
		os.Exit(1)
	}

	//Pre-Step
	switch *action {
	case "scale":
		log.Println("Pushing the app outside measurment time")
		must("pushing app", cf.Push(appName, *appDir, *stack, *buildpack, *startCommand))
		appGuid, err = cf.AppGuid(appName)
	}

	//Start Firehose
	log.Println("Buffering all messages from Firehose in the background.")
	firehoseEvents := make([]*events.Envelope, 100)
	cnsmr := consumer.New(*dopplerAddress, &tls.Config{InsecureSkipVerify: true}, nil)
	defer cnsmr.Close()
	stopFirehose := make(chan struct{})
	msgChan, errChan := cnsmr.Firehose("cfbench", string(authToken))
	go func(stop <-chan struct{}, msg <-chan *events.Envelope, err <-chan error) {
		for {
			select {
			case msg := <-msg:
				firehoseEvents = append(firehoseEvents, msg)
			case err := <-err:
				mustNot("consuming firehose", err)
			case <-stop:
				return
			}
		}
	}(stopFirehose, msgChan, errChan)

	log.Println("Waiting a few seconds to verify messages are being recorded")
	time.Sleep(time.Second * 5)

	//Benchmark Part
	var phases bench.Phases
	switch *action {
	case "push":
		must("pushing app", cf.Push(appName, *appDir, *stack, *buildpack, *startCommand))
		appGuid, err = cf.AppGuid(appName)
		mustNot("getting app GUID", err)
		phases = bench.ExtractBenchmarkPush(appGuid, instances)
	case "scale":
		appGuid, err = cf.AppGuid(appName)
		err := cf.Scale(appName, instances)
		mustNot("scaling app", err)
		phases = bench.ExtractBenchmarkScale(appGuid, instances)
	}

	log.Println("Waiting a few seconds in case some relevant messages are late")
	time.Sleep(time.Second * 5)

	//Close Firehose and process
	close(stopFirehose)
	log.Printf("Results:\n")
	phases.PopulateTimestamps(appGuid, firehoseEvents)

	//Print Results
	for _, phase := range phases {
		if phase.IsValid() {
			log.Printf("%s: %s (%s - %s)\n", phase.Name, phase.Duration().String(),
				time.Unix(0, phase.StartTimestamp), time.Unix(0, phase.EndTimestamp))
		} else {
			log.Printf("%s: %s (%s - %s)\n", phase.Name, "invalid measurement",
				time.Unix(0, phase.StartTimestamp), time.Unix(0, phase.EndTimestamp))
		}
	}

	//Clean up
	must("deleting app", cf.Delete(appName))
	must("purge routes", cf.PurgeRoutes())

	if jsonOutput {
		jsonResult := datadog.BuildJSONOutput(phases, tags)
		err = json.NewEncoder(os.Stdout).Encode(jsonResult)
		if err != nil {
			panic(err)
		}
	}
}

func envMustHave(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("please set %s\n", key)
		os.Exit(1)
	}
	return value
}

func mustNot(action string, err error) {
	if err != nil {
		log.Printf("error %s: %s\n", action, err)
		os.Exit(1)
	}
}

var must = mustNot

type tagList []string

func (p *tagList) String() string {
	return fmt.Sprintf("%v", *p)
}

func (p *tagList) Set(tag string) error {
	if tag == "" {
		return errors.New("Cannot set blank tag")
	}

	*p = append(*p, tag)
	return nil
}
