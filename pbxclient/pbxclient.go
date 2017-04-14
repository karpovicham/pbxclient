//pbxclient package connects to one of the the nats servers and makes out calls
package pbxclient

import (
	"errors"
	"github.com/nats-io/go-nats"
	"strings"
	"time"
)

//ClientConfig is config for PBX Client
//Servers are divided by comma, example: "nats://localhost:1222,nats://localhost:1223,nats://localhost:1224"
//provide self-signed certs if you use secure connection, example:
//Servers: "tls://nats.demo.io:4443"
//RootCAs: "conf/certs/ca.pem,conf/certs/ca2.pem"
type ClientConfig struct {
	Servers string
	RootCAs string
}

//Client stores NATS encoded conn and allows to make out call
type Client struct {
	nec natsEncodedConnector
}

type natsEncodedConn struct {
	*nats.EncodedConn
}

//natsEncodedConnector interface makes it possible to mock NATS encoded connection
type natsEncodedConnector interface {
	IsConnected() bool
	Close()
	Publish(subject string, v interface{}) error
	Request(subject string, v interface{}, vPtr interface{}, timeout time.Duration) error
	Subscribe(subject string, cb nats.Handler) (*nats.Subscription, error)
}

//OutCall has necessary data to make a request
type OutCall struct {
	CallTimeout      int    `json:"callTimeout"`
	CarrierID        int    `json:"carrierId"`
	Cmd              string `json:"cmd"`
	DestPhoneNumber  string `json:"destPhoneNumber"`
	Endpoint         string `json:"endpoint"`
	IGRP             int    `json:"igrp"`
	PBXClientTimeout int    `json:"pbxClientTimeout"`
	PBXHost          string `json:"pbxHost"`
	SrcPhoneNumber   string `json:"srcPhoneNumber"`
}

//OutCallResponse has out call response data
type OutCallResponse struct {
	ResponseStatus int
	ResponseData   map[string]string
}

var ErrConn = errors.New("Not connected")

//NewClient returns PBX Client with an opened NATS encoded connection
func NewClient(conf *ClientConfig) (*Client, error) {
	//Use secure connection if RootCAs (self-signed certs) are provided
	var options []nats.Option
	if conf.RootCAs != "" {
		rootCAsSlice := strings.Split(conf.RootCAs, ",")
		options = append(options, nats.RootCAs(rootCAsSlice...))
	}

	client := &Client{}

	//Set up connection
	nc, err := nats.Connect(conf.Servers, options...)
	if err != nil {
		return client, err
	}

	//Make sure we can connect to NATS server
	if !nc.IsConnected() {
		return client, ErrConn
	}

	//Get JSON encoded connection
	nec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		return client, err
	}

	client.nec = &natsEncodedConn{nec}
	return client, nil
}

//MakeOutCall sends a message via nats to the PBX server and wait for response
func (c *Client) MakeOutCall(oc *OutCall) (*OutCallResponse, error) {
	ocResp := &OutCallResponse{}

	//Make sure NATS server is connected
	if !c.IsConnected() {
		return nil, ErrConn
	}

	//Perform a Request(subject, req data, resp data, timeout) call with the Inbox reply for the data.
	//A response will be decoded into the vPtrResponse.
	if err := c.nec.Request(oc.PBXHost, oc, ocResp, time.Duration(oc.PBXClientTimeout)*time.Second); err != nil {
		return nil, err
	}

	return ocResp, nil
}

//Close func closes client nats connection
func (c *Client) Close() {
	if c.nec != nil {
		c.nec.Close()
	}
}

//IsConnected func checks if NATS server is connected, note that client can be set without NATS connection
func (c *Client) IsConnected() bool {
	if c.nec != nil {
		return c.nec.IsConnected()
	}

	return false
}

//IsConnected func expand nats.EncodedConn to matches natsEncodedConnector interface
func (nec *natsEncodedConn) IsConnected() bool {
	return nec.Conn.IsConnected()
}

//Request func expand nats.EncodedConn to check also if Connection is Opened
func (nec *natsEncodedConn) Request(subject string, v interface{}, vPtr interface{}, timeout time.Duration) error {
	if nec == nil || !nec.Conn.IsConnected() {
		return ErrConn
	}

	if err := nec.EncodedConn.Request(subject, v, vPtr, timeout); err != nil {
		return err
	}

	return nil
}
