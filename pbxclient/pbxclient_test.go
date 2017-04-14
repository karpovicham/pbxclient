//pbxclient tests with mocked NATS connection
package pbxclient

import (
	"fmt"
	"github.com/infinitytracking/icc-go/status"
	"github.com/nats-io/go-nats"
	"reflect"
	"testing"
)

//sprintError prints error message with expected and original results, it makes errors more readable in console
func sprintError(message string, expected interface{}, result interface{}) string {
	return fmt.Sprintf("%s:\n"+
		"- expected: '%v'\n"+
		"- result:   '%v'",
		message, expected, result,
	)
}

//TestClient func test Client and MakeOutCall function with non exists server
func TestClient(t *testing.T) {
	pbxClient, err := NewClient(&ClientConfig{Servers: "nats://test:1111"})
	expectedErr := nats.ErrNoServers
	if err != expectedErr {
		t.Error(sprintError("Failed to get PBX client with wrong server", expectedErr, err))
	}

	oc := &OutCall{
		CallTimeout:      5,
		CarrierID:        1,
		Cmd:              "test_numbertesting",
		DestPhoneNumber:  "+441711112222",
		Endpoint:         "test_endpoint",
		IGRP:             1,
		PBXClientTimeout: 6,
		PBXHost:          "test_pbx.host",
		SrcPhoneNumber:   "+441712341234",
	}

	_, err = pbxClient.MakeOutCall(oc)
	expectedErr = ErrConn
	if err != expectedErr {
		t.Error(sprintError("Failed to make out call with not connected NATS server", expectedErr, err))
	}
}

//TestClient_MakeOutCall func makes out call using mocked NATS connection
func TestClient_MakeOutCall(t *testing.T) {
	pbxClient, _ := NewMockClient()

	oc := &OutCall{
		CallTimeout:      5,
		CarrierID:        1,
		Cmd:              "test_numbertesting",
		DestPhoneNumber:  "+441711112222",
		Endpoint:         "test_endpoint",
		IGRP:             1,
		PBXClientTimeout: 6,
		PBXHost:          "test_pbx.host",
		SrcPhoneNumber:   "+441712341234",
	}

	resultOcResp, err := pbxClient.MakeOutCall(oc)
	if err != nil {
		t.Fatal(sprintError("Error making out call", nil, err))
	}

	expectedOcResp := &OutCallResponse{
		ResponseStatus: status.OK,
		ResponseData: map[string]string{
			"dtmfReturned":   "441711112222",
			"hangupCause":    "NORMAL_CLEARING",
			"hangupSipCode":  "sip:200",
			"hangupQ850Code": "16",
		},
	}

	if !reflect.DeepEqual(expectedOcResp, resultOcResp) {
		t.Error(sprintError("Failed to get correct Out call response", expectedOcResp, resultOcResp))
	}
}

//TestClient_MakeOutCall_withTimeoutError func makes out call using mocked NATS connection with timeout error
func TestClient_MakeOutCall_withTimeoutError(t *testing.T) {
	pbxClient, _ := NewMockClient()

	oc := &OutCall{
		CallTimeout:      5,
		CarrierID:        1,
		Cmd:              "test_numbertesting",
		DestPhoneNumber:  "+441711112222",
		Endpoint:         "test_endpoint",
		IGRP:             1,
		PBXClientTimeout: 3,
		PBXHost:          "test_pbx.host",
		SrcPhoneNumber:   "+441712341234",
	}

	_, err := pbxClient.MakeOutCall(oc)

	expectedErr := nats.ErrTimeout
	if err == nil || err != expectedErr {
		t.Error(sprintError("Error making out call with timeout error", expectedErr, err))
	}
}
