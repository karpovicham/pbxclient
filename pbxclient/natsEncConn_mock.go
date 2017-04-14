//Provides mocked NATS encoded connection for pbxclient tests
package pbxclient

import (
	"errors"
	"fmt"
	"github.com/infinitytracking/icc-go/status"
	"github.com/nats-io/go-nats"
	"strings"
	"time"
)

//natsEncConnMock struct is a mock nats.EncodedConn struct which implements natsEncodedConnector interface
type natsEncConnMock struct {
}

//NewMockClient returns pbxclient with a mocked NATS encoded connection
func NewMockClient() (*Client, error) {
	return &Client{nec: &natsEncConnMock{}}, nil
}

//Request func mocks nats encoded conn Request func
func (ncm *natsEncConnMock) Request(subject string, v interface{}, vPtr interface{}, timeout time.Duration) error {
	defaultTimeout := 5 * time.Second

	//Sleep for a provided time in case if it's shorter than the default one and return NATS timeout error
	if timeout < defaultTimeout {
		time.Sleep(timeout)
		return nats.ErrTimeout
	}

	time.Sleep(defaultTimeout)

	switch vType := v.(type) {
	case *OutCall:
		//Set success Out call response
		ocResp := vPtr.(*OutCallResponse)
		ocResp.ResponseStatus = status.OK
		ocResp.ResponseData = map[string]string{
			"dtmfReturned":   strings.Replace(vType.DestPhoneNumber, "+", "", -1),
			"hangupCause":    "NORMAL_CLEARING",
			"hangupSipCode":  "sip:200",
			"hangupQ850Code": "16",
		}
		return nil
	default:
		return errors.New(fmt.Sprint("Unsupported request data type: ", vType))
	}
}

//Close func is a mocked nats Close func
func (ncm *natsEncConnMock) Close() {
}

//Subscribe func mocks nats encoded conn Subscribe func
func (ncm *natsEncConnMock) Subscribe(subject string, cb nats.Handler) (*nats.Subscription, error) {
	return &nats.Subscription{}, nil
}

//Publish func mocks nats encoded conn Publish func
func (ncm *natsEncConnMock) Publish(subject string, v interface{}) error {
	return nil
}

//IsConnected func mocks nats encoded conn IsConnected func
func (ncm *natsEncConnMock) IsConnected() bool {
	return true
}
