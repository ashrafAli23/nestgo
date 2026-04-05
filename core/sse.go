package core

import (
	"fmt"
	"io"
	"strings"
)

// SSEEvent represents a Server-Sent Event.
type SSEEvent struct {
	// Event is the event type (optional). Maps to the "event:" field.
	Event string
	// Data is the event payload. Maps to the "data:" field.
	Data string
	// ID is the event ID (optional). Maps to the "id:" field.
	ID string
	// Retry is the reconnection time in milliseconds (optional). Maps to the "retry:" field.
	Retry int
}

// Format returns the SSE-formatted string for this event.
func (e SSEEvent) Format() string {
	var b strings.Builder
	if e.ID != "" {
		fmt.Fprintf(&b, "id: %s\n", e.ID)
	}
	if e.Event != "" {
		fmt.Fprintf(&b, "event: %s\n", e.Event)
	}
	if e.Retry > 0 {
		fmt.Fprintf(&b, "retry: %d\n", e.Retry)
	}
	fmt.Fprintf(&b, "data: %s\n\n", e.Data)
	return b.String()
}

// SSEStream represents a channel-based SSE stream.
// Create one, start a goroutine to send events, then pass it to SSE().
// Close the channel when done.
type SSEStream chan SSEEvent

// NewSSEStream creates a new SSE stream with the given buffer size.
func NewSSEStream(bufferSize int) SSEStream {
	return make(SSEStream, bufferSize)
}

// sseReader adapts an SSEStream to io.Reader for use with SendStream.
type sseReader struct {
	stream SSEStream
	buf    []byte
}

func (r *sseReader) Read(p []byte) (n int, err error) {
	// Drain buffer first
	if len(r.buf) > 0 {
		n = copy(p, r.buf)
		r.buf = r.buf[n:]
		return n, nil
	}

	// Read next event from channel
	event, ok := <-r.stream
	if !ok {
		return 0, io.EOF
	}

	data := []byte(event.Format())
	n = copy(p, data)
	if n < len(data) {
		r.buf = data[n:]
	}
	return n, nil
}

// SSE sets up Server-Sent Events headers and streams events from the channel.
// The function blocks until the channel is closed or the client disconnects.
//
//	func (ctrl *Controller) Events(c core.Context) error {
//	    stream := core.NewSSEStream(10)
//	    go func() {
//	        defer close(stream)
//	        for i := 0; i < 5; i++ {
//	            stream <- core.SSEEvent{
//	                Event: "message",
//	                Data:  fmt.Sprintf(`{"count": %d}`, i),
//	                ID:    fmt.Sprintf("%d", i),
//	            }
//	            time.Sleep(time.Second)
//	        }
//	    }()
//	    return core.SSE(c, stream)
//	}
func SSE(c Context, stream SSEStream) error {
	c.SetHeader("Content-Type", "text/event-stream")
	c.SetHeader("Cache-Control", "no-cache")
	c.SetHeader("Connection", "keep-alive")
	c.SetHeader("X-Accel-Buffering", "no")

	reader := &sseReader{stream: stream}
	return c.SendStream(reader)
}
