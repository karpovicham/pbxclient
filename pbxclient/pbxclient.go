//Package pbxclient connects to one of the the nats servers and makes out calls
package pbxclient

import (
	"errors"
	"fmt"
	"github.com/nats-io/go-nats"
	"log"
	"time"
)

//Client stores nats conn and its config
type Client struct {
	config *ClientConfig
	nc     *nats.Conn
}

//ClientConfig is a config to set up nats connection
//servers are divided by comma example "nats://localhost:1222,nats://localhost:1223,nats://localhost:1224"
//set RootCAs if you use secure connection, for example:
//Servers: "tls://nats.demo.io:4443"
//RootCAs: "./conf/certs/ca.pem"
type ClientConfig struct {
	Servers  string
	RootCAs  string
}

//OutCall has necessary data to make a request
type OutCall struct {
	CallerId    int
	CarrierId   int
	PhoneNumber string
	Igrp        int
	Endpoint    string
	PbxHost     string
}

//OutCallResponse has out call response data
type OutCallResponse struct {
	ResponseStatus int
	ResponseData   map[string]string
}

//NewClient returns pbxclient with set up connection
func NewClient(config *ClientConfig) (*Client, error) {
	client := &Client{config: config}

	//Try to connect to the nats server using configs
	if err := client.setUpNatsConn(); err != nil {
		return nil, err
	}
	defer client.nc.Close()

	return client, nil
}

//setUpNatsConn check what connection we should do depending on provided configs and sets up connection
func (c *Client) setUpNatsConn() (err error) {
	//TODO: set up default options
	options := []nats.Option{}

	//Use secure connection if we have a CA
	if c.config.RootCAs != "" {
		options = append(options, nats.RootCAs(c.config.RootCAs))
	}

	//Connect to the nats server and store opened connection in the client
	c.nc, err = nats.Connect(c.config.Servers, options...)
	if err != nil {
		return err
	}

	return nil
}

//ensureConn checks if connection is active, if not - try to reconnect
//looks like nats api has its own reconn feature, we will test it once nats server is set up and available for testing
func (c *Client) ensureConn() error {
	if !c.nc.IsConnected() {
		//TODO: Reconnect
		return nil
	}

	return errors.New("Could not set up connection")
}

//MakeOutCall sends a message via nats to the PBX server and wait for response
func (c *Client) MakeOutCall(oc *OutCall) (*OutCallResponse, error) {
	c.logInfo("Started making a call")
	defer c.logInfo("Finished making a call")

	//Ensure connection is opened
	if err := c.ensureConn(); err != nil {
		return nil, err
	}

	//TODO: replace sleep with making actual NATS call
	time.Sleep(5 * time.Second)

	ocResp := &OutCallResponse{
		ResponseStatus: 200,
		ResponseData:   map[string]string{"dtmfReturned": "123456789"},
	}

	return ocResp, nil
}

//logInfo logs info message
func (c *Client) logInfo(a ...interface{}) {
	log.Print("pbxclient: servers:", c.config.Servers, ", Info: ", fmt.Sprint(a...))
}

//logInfo logs error message
func (c *Client) logError(a ...interface{}) {
	log.Print("pbxclient: servers:", c.config.Servers, ", Error: ", fmt.Sprint(a...))
}
