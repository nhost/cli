package nhtracing

import (
	"net/http"

	"github.com/google/uuid"
)

const (
	headerTraceID      = "X-B3-TraceId"
	headerSpanID       = "X-B3-SpanId"
	headerParentSpanID = "X-B3-ParentSpanId"
)

type Trace struct {
	TraceID      string
	ParentSpanID string
	SpanID       string
}

// NewTrace creates a new trace with a new `TraceID`.
func NewTrace() Trace {
	return Trace{
		TraceID:      NewID(),
		ParentSpanID: "",
		SpanID:       "",
	}
}

// NewSpan create a new trace with the same `TraceID`, `ParentSpanID` set as the current `SpanID`
// and a new `SpanID`.
func (t Trace) NewSpan() Trace {
	return Trace{
		TraceID:      t.TraceID,
		ParentSpanID: t.SpanID,
		SpanID:       NewID(),
	}
}

// FromHTTPHeaders extracts tracing information from HTTP headers.
// If no tracing information is found, a new trace is created with only `TraceID` set.
func FromHTTPHeaders(headers http.Header) Trace {
	traceID := headers.Get(headerTraceID)
	if traceID == "" {
		traceID = uuid.New().String()
	}

	spanID := headers.Get(headerSpanID)
	parentSpanID := headers.Get(headerParentSpanID)

	return Trace{
		TraceID:      traceID,
		ParentSpanID: parentSpanID,
		SpanID:       spanID,
	}
}

// ToHTTPHeaders adds tracing information to HTTP headers.
func ToHTTPHeaders(trace Trace, header http.Header) {
	header.Set(headerTraceID, trace.TraceID)
	header.Set(headerParentSpanID, trace.ParentSpanID)
	header.Set(headerSpanID, trace.SpanID)
}

func NewID() string {
	return uuid.New().String()
}
