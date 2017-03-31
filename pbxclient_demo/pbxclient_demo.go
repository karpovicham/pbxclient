//Demo application to work with a pbxclient
package main

import (
	"context"
	"fmt"
	"github.com/infinitytracking/icc-go/pbxclient"
	"github.com/nats-io/go-nats"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const callTimeout = 6 * time.Second

type numberData struct {
	phoneNumber string
	callerId    int
	carrierId   int
	igrp        int
	endpoint    string
	pbxHost     string
}

var wg sync.WaitGroup

var numbersData = []numberData{
	{phoneNumber: "11111", callerId: 1, carrierId: 1, igrp: 1, endpoint: "demo", pbxHost: "demo"},
	{phoneNumber: "11112", callerId: 1, carrierId: 1, igrp: 1, endpoint: "demo", pbxHost: "demo"},
	{phoneNumber: "11113", callerId: 1, carrierId: 1, igrp: 1, endpoint: "demo", pbxHost: "demo"},
}

func main() {
	log.Print("Started demo application")

	//Setup context for cancelling
	ctx, cancel := context.WithCancel(context.Background())
	go signalHandler(cancel)

	config := &pbxclient.ClientConfig{
		Servers: nats.DefaultURL,
	}

	client, err := pbxclient.NewClient(config)
	if err != nil {
		log.Print("Error setting up pbxclient")
		return
	}

	//Create a separate process for each number
	//Use wg sync to control gourutines for Demo application (it's simpler and less code)
	for _, numberData := range numbersData {
		wg.Add(1)
		go numberData.processNumber(ctx, client)
	}

	log.Print("Waiting for all processes to be finished")
	wg.Wait()
	log.Print("Done! Close application")
}

//Process one number, use pbx client to make a call and wait for the call response
func (nd numberData) processNumber(ctx context.Context, client *pbxclient.Client) {
	defer wg.Done()

	nd.logInfo("Process started")
	defer nd.logInfo("Process finished")

	oc := &pbxclient.OutCall{
		CallerId:    nd.callerId,
		CarrierId:   nd.carrierId,
		PhoneNumber: nd.phoneNumber,
		Igrp:        nd.igrp,
		Endpoint:    nd.endpoint,
		PbxHost:     nd.pbxHost,
	}

	respCh := make(chan *pbxclient.OutCallResponse)
	//TODO: use context or channel to close this func if we got timeout error
	go func() {
		ocResp, err := client.MakeOutCall(oc)
		if err != nil {
			nd.logError(err)
			return
		}

		respCh <- ocResp
	}()

	//Wait for call response
	//Exit in case if we have timeout error or the process is stopped by signal
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(callTimeout):
			nd.logError("Got timeout, exit process")
			return
		case pbxCallResp := <-respCh:
			//do something with pbxCallResp
			nd.logInfo("Success, repsonse status: ", pbxCallResp.ResponseStatus)
			return
		}
	}
}

func (nd numberData) logInfo(a ...interface{}) {
	log.Print("Demo: Phone number ", nd.phoneNumber, ", Info: ", fmt.Sprint(a...))
}
func (nd numberData) logError(a ...interface{}) {
	log.Print("Demo: Phone number ", nd.phoneNumber, ", Error: ", fmt.Sprint(a...))
}

func signalHandler(cancel context.CancelFunc) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	for {
		<-sigc
		log.Print("Got signal, cancelling context")
		cancel()
	}
}
