package applog

import (
	"bytes"
	"context"
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestBasicLog(t *testing.T) {
	buf := new(bytes.Buffer)

	type ContextKey string
	key := ContextKey("traceid")

	logrus.SetOutput(buf)
	logrus.SetFormatter(&Formatter{TraceKey: key})
	logrus.SetReportCaller(true)

	ctx := context.WithValue(context.Background(), key, "abcdef123456")

	logrus.WithContext(ctx).WithFields(
		logrus.Fields{
			"animal": "walrus",
			"age":    12,
		}).Info("My message")

	// Get the current source code line and step back 3 to get the line where the logging happened
	logLine := currentLine() - 3

	// Convert the log output back into a googleLogEntry struct
	e := outputToGoogleLogEntry(buf, t)

	if e.Message != "My message" {
		t.Errorf("Wanted message='%s', got '%s'", "My message", e.Message)
	}

	if e.Severity != "info" {
		t.Errorf("Wanted severity='%s', got '%s'", "info", e.Severity)
	}

	if e.Additional["animal"] != "walrus" {
		t.Errorf("Wanted animal='walrus', got '%s'", e.Additional["animal"])
	}

	if e.Additional["age"] != 12.0 {
		t.Errorf("Wanted age=12, got %v", e.Additional["age"])
	}

	if e.TraceID != "abcdef123456" {
		t.Errorf("Wanted traceid='abcdef123456', got '%s'", e.TraceID)
	}

	if e.Type != "" {
		t.Errorf("Wanted @type='', got '%s'", e.Type)
	}

	if e.SourceLocation.File != "formatter_test.go" {
		t.Errorf("Wanted source file ='formatter_test.go', got '%s'\n", e.SourceLocation.File)
	}

	if e.SourceLocation.Line != logLine {
		t.Errorf("Wanted line = %d, got %d\n", logLine, e.SourceLocation.Line)
	}
}

// Test various conditions of the context, context key and value stored in the context for the trace ID
// to make sure that no panics occur
func TestNoTraceID(t *testing.T) {
	buf := new(bytes.Buffer)

	logrus.SetOutput(buf)
	logrus.SetFormatter(&Formatter{})
	logrus.SetReportCaller(false)

	// Context is NIL, trace key is NIL
	logrus.WithContext(nil).Info("Hello")
	if strings.TrimSpace(buf.String()) != `{"message":"Hello","severity":"info"}` {
		t.Errorf("Test 1: wanted [%s], got [%s]", `{"message":"Hello","severity":"info"}`, strings.TrimSpace(buf.String()))
	}
	buf.Reset()

	// Context is OK, Trace key is NIL
	logrus.WithContext(context.Background()).Info("Hello")
	if strings.TrimSpace(buf.String()) != `{"message":"Hello","severity":"info"}` {
		t.Errorf("Test 2: wanted [%s], got [%s]", `{"message":"Hello","severity":"info"}`, strings.TrimSpace(buf.String()))
	}
	buf.Reset()

	type ContextKey string
	key := ContextKey("traceid")
	logrus.SetFormatter(&Formatter{TraceKey: key})

	// Context is OK, Trace key is OK, value is NIL
	ctx := context.WithValue(context.Background(), key, nil)

	logrus.WithContext(ctx).Info("Hello")
	if strings.TrimSpace(buf.String()) != `{"message":"Hello","severity":"info"}` {
		t.Errorf("Test 2: wanted [%s], got [%s]", `{"message":"Hello","severity":"info"}`, strings.TrimSpace(buf.String()))
	}
	buf.Reset()

	ctx = context.WithValue(context.Background(), key, 100)

	// Context is OK, Trace key is OK, value is INT
	logrus.WithContext(ctx).Info("Hello")
}

func TestTraceID(t *testing.T) {
	buf := new(bytes.Buffer)

	type ContextKey string
	key := ContextKey("traceid")

	logrus.SetOutput(buf)
	logrus.SetFormatter(&Formatter{TraceKey: key})

	ctx := context.WithValue(context.Background(), key, 100)

	// Value is INT (should not be logged)
	logrus.WithContext(ctx).Info("Hello")
	if strings.TrimSpace(buf.String()) != `{"message":"Hello","severity":"info"}` {
		t.Errorf("Test 2: wanted [%s], got [%s]", `{"message":"Hello","severity":"info"}`, strings.TrimSpace(buf.String()))
	}
	buf.Reset()

	ctx = context.WithValue(context.Background(), key, "ABC")

	// Value is STRING
	logrus.WithContext(ctx).Info("Hello")
	if strings.TrimSpace(buf.String()) != `{"message":"Hello","severity":"info","logging.googleapis.com/trace":"ABC"}` {
		t.Errorf("Test 2: wanted [%s], got [%s]", `{"message":"Hello","severity":"info","logging.googleapis.com/trace":"ABC"}`, strings.TrimSpace(buf.String()))
	}

}

func TestNoCaller(t *testing.T) {
	buf := new(bytes.Buffer)

	logrus.SetOutput(buf)
	logrus.SetFormatter(&Formatter{})
	logrus.SetReportCaller(false) // turn off reporting caller

	logrus.WithFields(
		logrus.Fields{
			"animal": "walrus",
			"age":    12,
		}).Info("My message")

	e := outputToGoogleLogEntry(buf, t)

	if e.SourceLocation != nil {
		t.Errorf("Wanted SourceLocation = nil, got %+v", e.SourceLocation)
	}

}

// Convert a logged JSON string to a Google Log Entry struct
func outputToGoogleLogEntry(buf *bytes.Buffer, t *testing.T) googleLogEntry {
	var e googleLogEntry
	err := json.Unmarshal(buf.Bytes(), &e)
	if err != nil {
		t.Errorf("couldn't convert json to log entry: %s", err.Error())
	}
	return e
}

// Get the source code line this function was called from
func currentLine() int {
	if _, _, l, ok := runtime.Caller(1); ok {
		return l
	}
	return 0
}
