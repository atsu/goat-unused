package health

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

const SignalUrl = "https://signal.atsu.io/"

// Signal is a mechanism for sending a health.Event and Data (interface{})
type Signal interface {
	Report(Event, interface{}) error

	RawReport([]byte) error
}

// HttpSignal implements against our canonical endpoint SignalUrl
type HttpSignal struct {
	client *http.Client
	url    string
}

func NewHttpSignal(url string) *HttpSignal {
	c := &http.Client{
		Timeout: time.Second * 10,
	}
	return &HttpSignal{client: c, url: url}
}

// Report will Marshal and send, or error
func (r *HttpSignal) Report(e Event, i interface{}) error {
	e.Data = i

	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	return r.RawReport(b)
}

// RawReport will POST the provided jsonPayload to SignalUrl
func (r *HttpSignal) RawReport(jsonPayload []byte) error {
	_, err := r.client.Post(r.url, "application/json", bytes.NewBuffer(jsonPayload))
	return err
}
