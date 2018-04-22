package sse_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/jadr2ddude/sse"
)

func TestScannedEventDecoding(t *testing.T) {
	cc := []struct {
		name  string
		input string
		event sse.ScannedEvent
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
			event: sse.ScannedEvent{Data: "ok\n"},
		},
		{
			name:  "one data line with retry",
			input: "data:ok\nretry:15\n\n",
			event: sse.ScannedEvent{Data: "ok\n", Retry: 15, RetrySet: true},
		},
		{
			name:  "one data line with zero retry",
			input: "data:ok\nretry: 0\n\n",
			event: sse.ScannedEvent{Data: "ok\n", RetrySet: true},
		},
		{
			name:  "one data line with type",
			input: "event: type\ndata:ok\n\n",
			event: sse.ScannedEvent{Type: "type", Data: "ok\n"},
		},
		{
			name:  "one data line with id",
			input: "data:ok\nid:1\n\n",
			event: sse.ScannedEvent{ID: "1", IDSet: true, Data: "ok\n"},
		},
		{
			name:  "one data line with empty id",
			input: "data:ok\nid\n\n",
			event: sse.ScannedEvent{IDSet: true, Data: "ok\n"},
		},
		{
			name:  "U+0000 in ID",
			input: "id:\0001\n\n",
			event: sse.ScannedEvent{},
		},
		{
			name:  "one data line with leading space",
			input: "data: ok\n\n",
			event: sse.ScannedEvent{Data: "ok\n"},
		},
		{
			name:  "one data line with two leading spaces",
			input: "data:  ok\n\n",
			event: sse.ScannedEvent{Data: " ok\n"},
		},
		{
			name:  "comment at the beginning",
			input: ":some comment\ndata:ok\n\n",
			event: sse.ScannedEvent{Data: "ok\n"},
		},
		{
			name:  "comment at the end",
			input: "data:ok\n:some comment\n\n",
			event: sse.ScannedEvent{Data: "ok\n"},
		},
		{
			name:  "empty data",
			input: "data:\n\n",
			event: sse.ScannedEvent{Data: "\n"},
		},
		{
			name:  "empty data (without ':')",
			input: "data\n\n",
			event: sse.ScannedEvent{Data: "\n"},
		},
		{
			name:  "multiple data lines",
			input: "data:1\ndata: 2\ndata:3\n\n",
			event: sse.ScannedEvent{Data: "1\n2\n3\n"},
		},
	}

	for _, c := range cc {
		t.Run(c.name, func(t *testing.T) {
			scanner := sse.NewScanner(bytes.NewBufferString(c.input))
			e, err := scanner.Event()
			if err != c.err {
				t.Errorf("got error '%v', expected '%v'", err, c.err)
			}
			if e != c.event {
				t.Errorf("got %#v, expected %#v", e, c.event)
			}
		})
	}
}
