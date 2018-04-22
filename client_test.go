package sse

import (
	"bytes"
	"io"
	"testing"

	"github.com/jadr2ddude/sse"
)

func TestEventDecoding(t *testing.T) {
	cc := []struct {
		name  string
		input string
		event Event
		err   error
	}{
		{
			name:  "empty_event",
			input: "",
			err:   io.EOF,
		},
		{
			name:  "incomplete_event",
			input: "data:ok\n",
			err:   io.EOF,
		},
		{
			name:  "one data line",
			input: "data:ok\n\n",
			event: sse.Event{Name: "message", Data: "ok"},
		},
		{
			name:  "one data line with leading space",
			input: "data: ok\n\n",
			event: sse.Event{Name: "message", Data: "ok"},
		},
		{
			name:  "one data line with two leading spaces",
			input: "data:  ok\n\n",
			event: sse.Event{Name: "message", Data: " ok"},
		},
		{
			name:  "comment at the beginning",
			input: ":some comment\ndata:ok\n\n",
			event: sse.Event{Name: "message", Data: "ok"},
		},
		{
			name:  "comment at the end",
			input: "data:ok\n:some comment\n\n",
			event: sse.Event{Name: "message", Data: "ok"},
		},
		{
			name:  "empty data",
			input: "data:\n\n",
			event: sse.Event{Name: "message", Data: ""},
		},
		{
			name:  "empty data (without ':')",
			input: "data\n\n",
			event: sse.Event{Name: "message", Data: ""},
		},
		{
			name:  "multiple data lines",
			input: "data:1\ndata: 2\ndata:3\n\n",
			event: sse.Event{Name: "message", Data: "1\n2\n3"},
		},
	}

	for _, c := range cc {
		t.Run(c.name, func(t *testing.T) {
			client := NewClient(bytes.NewBufferString(c.input))
			e, err := client.Event()
			if err != c.err {
				t.Errorf("got error '%v', expected '%v'", err, c.err)
			}
			if e != c.event {
				t.Errorf("got %#v, expected %#v", e, c.event)
			}
			err = client.Close()
			if err != nil {
				t.Errorf("could not close the client: %v", err)
			}
		})
	}
}
