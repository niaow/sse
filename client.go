package sse

import (
	"errors"
	"io"
	"net/http"
)

//Client is a SSE client
type Client struct {
	s           *Scanner
	lastEventID string
	close       func() error
}

// NewClient returns a client which will parse Event from the io.Reader
func NewClient(r io.Reader) *Client {
	s := NewScanner(r)
	var close func() error
	closer, ok := (r).(io.Closer)
	if ok {
		close = closer.Close
	}
	return &Client{s: s, close: close}
}

//ErrClosedClient is an error indicating that the client has been used after it was closed.
var ErrClosedClient = errors.New("client closed")

// Event reads an event from the stream.
func (c *Client) Event() (ev Event, err error) {
	if c.s == nil {
		return Event{}, ErrClosedClient
	}

	var event ScannedEvent
	// Wait for an event with some data
	for event.Data == "" {
		event, err = c.s.Event()
		if err != nil {
			return Event{}, err
		}
	}
	if event.Type == "" {
		ev.Name = "message"
	} else {
		ev.Name = event.Type
	}
	if event.Data != "" {
		// strip last \n
		ev.Data = event.Data[:len(event.Data)-1]
	}
	if event.IDSet {
		c.lastEventID = event.ID
	}
	// todo: set lastEventID of event

	return ev, nil
}

//Close closes the client
func (c *Client) Close() error {
	c.s = nil
	if c.close != nil {
		return c.close()
	}
	return nil
}

//ErrNotSSE is an error returned when a client recieves a non-SSE response
var ErrNotSSE = errors.New("content type is not 'text/event-stream'")

//Connect performs an SSE request and returns a Client.
func Connect(client *http.Client, request *http.Request) (*Client, error) {
	request.Header.Set("Accept", "text/event-stream")
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		resp.Body.Close()
		return nil, ErrNotSSE
	}
	return NewClient(resp.Body), nil
}
