//Demo application to work with a pbxclient
package main

import (
	"context"
	"fmt"
	"github.com/infinitytracking/icc-go/pbxclient"
	"github.com/nats-io/go-nats"
	"log"
	"log/syslog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type numberData struct {
	phoneNumber string
	callerId    string
	carrierId   int
	igrp        int
	endpoint    string
	pbxHost     string
}

var wg sync.WaitGroup

var numbersData = []numberData{
	{phoneNumber: "11111", callerId: "+1234", carrierId: 1, igrp: 1, endpoint: "demo", pbxHost: "demo"},
	{phoneNumber: "11112", callerId: "+1234", carrierId: 1, igrp: 1, endpoint: "demo", pbxHost: "demo"},
	{phoneNumber: "11113", callerId: "+1234", carrierId: 1, igrp: 1, endpoint: "demo", pbxHost: "demo"},
}

const appName = "PBX client demo app"

func main() {
	setUpSyslog(appName)
	log.Print("Started ", appName)
	defer log.Print("Finished ", appName)

	//Setup context for cancelling
	ctx, cancel := context.WithCancel(context.Background())
	go signalHandler(cancel)

	pbxClient, err := pbxclient.NewClient(&pbxclient.ClientConfig{Servers: nats.DefaultURL})
	if err != nil {
		log.Print("Error initialising pbxclient", err)
		return
	}
	defer pbxClient.Close()

	//Create a separate process for each number
	//Use wg sync to control goroutines for Demo application (it's simpler and less code)
	for _, numberData := range numbersData {
		wg.Add(1)
		go numberData.processNumber(ctx, pbxClient)
	}

	log.Print("Waiting for all processes to be finished")
	wg.Wait()
}

//setUpSyslog sets up syslog with a application name
func setUpSyslog(tag string) {
	log.SetFlags(0)
	syslogWriter, err := syslog.New(syslog.LOG_INFO, tag)
	if err == nil {
		log.SetOutput(syslogWriter)
	}
}

//logInfo prints info message
func (nd numberData) logInfo(a ...interface{}) {
	log.Print("Phone number ", nd.phoneNumber, ", Info: ", fmt.Sprint(a...))
}

//logError prints error message
func (nd numberData) logError(a ...interface{}) {
	log.Print("Error: Phone number ", nd.phoneNumber, ", Error: ", fmt.Sprint(a...))
}

//signalHandler function responds to SIGUP, SIGTERM signals to cleanly shutdown worker using context
func signalHandler(cancel context.CancelFunc) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	for {
		<-sigc
		log.Print("Got signal, cancelling context")
		cancel()
	}
}

//Process one number, use pbx client to make a call and wait for the call response
func (nd numberData) processNumber(ctx context.Context, client *pbxclient.Client) {
	defer wg.Done()

	nd.logInfo("Process started")
	defer nd.logInfo("Process finished")

	oc := &pbxclient.OutCall{
		DestPhoneNumber: nd.phoneNumber,
		CarrierID:       nd.carrierId,
		IGRP:            nd.igrp,
		Endpoint:        nd.endpoint,
		PBXHost:         nd.pbxHost,
		SrcPhoneNumber:  nd.callerId,
	}

	ocResp, err := client.MakeOutCall(oc)
	if err != nil {
		nd.logError("Making out call: ", err)
	}

	nd.logInfo("Success, response status: ", ocResp.ResponseStatus)
	return
}
