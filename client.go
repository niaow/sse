package sse

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"strings"
)

//Client is a SSE client
type Client struct {
	r io.ReadCloser
	s *bufio.Scanner
}

//ErrClosedClient is an error indicating that the client has been used after it was closed.
var ErrClosedClient = errors.New("client closed")

//ReadEvent reads an event from the stream.
func (c *Client) ReadEvent() (ev Event, err error) {
	//An exact implementation of https://www.w3.org/TR/2009/WD-eventsource-20090421/#event-stream-interpretations
	if c.s == nil {
		return Event{}, ErrClosedClient
	}
	for c.s.Scan() {
		ln := c.s.Text()
		if ln == "" {
			if ev.Data == "" || checkNCName(ev.Name) != nil {
				ev = Event{}
				continue
			}
			return
		}
		if strings.HasPrefix(ln, ":") { //comment
			continue
		}
		var key, val string
		if strings.Contains(ln, ":") {
			spl := strings.SplitN(ln, ":", 2)
			key, val = spl[0], spl[1]
			val = strings.TrimLeft(val, " ")
		} else {
			key = ln
		}
		switch key {
		case "event":
			ev.Name = val
		case "data":
			if val == "" {
				val = "\u000A"
			}
			ev.Data = val
		case "id": //unimplemented
		case "retry": //unimplemented
		default:
			//do nothing, as per the spec
		}
	}
	err = c.s.Err()
	return
}

//Close closes the client
func (c *Client) Close() error {
	c.s = nil
	return c.r.Close()
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
	return &Client{r: resp.Body, s: bufio.NewScanner(resp.Body)}, nil
}
