package pbxclient

import (
	"fmt"
	"github.com/nats-io/go-nats"
	"testing"
)

func TestNewClient(t *testing.T) {
	config := &ClientConfig{
		Servers: nats.DefaultURL,
	}

	pbxclient, err := NewClient(config)
	if err != nil {
		t.Fatal(err)
	}

	oc := &OutCall{
		CallerId:    1,
		CarrierId:   1,
		PhoneNumber: "12345",
		Igrp:        1,
		Endpoint:    "123",
		PbxHost:     "12345",
	}

	ocResp, err := pbxclient.MakeOutCall(oc)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(ocResp)

}
