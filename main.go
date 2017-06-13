package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"

	"code.cloudfoundry.org/cftrace/cflib"
	. "code.cloudfoundry.org/cftrace/process"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
)

func main() {
	dopplerAddress := envMustHave("DOPPLER_ADDR")
	authToken := envMustHave("CF_AUTH_TOKEN")
	appName := "randomX"
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	var appDir = flag.String("app-dir", pwd, "The directory of the app to push")
	flag.Parse()

	cf := cflib.Adapter{CfCliPath: "cf"}

	fmt.Println("=== Start recording all messages from Firehose in the background")
	msgBuffer := make([]*events.Envelope, 0, 200)
	cnsmr := consumer.New(dopplerAddress, &tls.Config{InsecureSkipVerify: true}, nil)
	defer cnsmr.Close()
	stop := make(chan interface{})
	go func(stop <-chan interface{}) {
		msgChan, errorChan := cnsmr.FilteredFirehose("whatisthat", string(authToken), consumer.LogMessages)
		go func() {
			for err := range errorChan {
				fmt.Fprintf(os.Stderr, "%v\n", err.Error())
			}
		}()

		for {
			select {
			case msg := <-msgChan:
				msgBuffer = append(msgBuffer, msg)
			case <-stop:
				fmt.Println("=== Stop recording messages from Firehose")
				return
			}
		}
	}(stop)

	if cf.Push(appName, *appDir) != nil {
		log.Fatal("cf push failed!")
	}

	appGuid, err := cf.AppGuid(appName)
	if err != nil {
		log.Fatal("cf app x --guid failed ")
	}

	if cf.Delete(appName) != nil {
		log.Fatal("cf delete failed!")
	}

	close(stop)
	p := NewPushProcess(appGuid)
	p.GetTimestamps(msgBuffer)
	p.PrintResult()
	//InvestigateMessages(msgBuffer, appGuid)

}

func envMustHave(key string) string {
	value := os.Getenv(key)
	str := `
export CF_AUTH_TOKEN=$(cf oauth-token); export DOPPLER_ADDR=$(cf curl /v2/info| jq -r .doppler_logging_endpoint)
	`
	if value == "" {
		fmt.Printf("please set envs \n%s\n", str)
		os.Exit(1)
	}
	return value
}
