package applog

import (
	"encoding/json"
	"path"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

type Formatter struct {
	TraceKey interface{}
}

type googleLogEntry struct {
	Message        string          `json:"message"`
	Severity       string          `json:"severity"`
	Additional     logrus.Fields   `json:"additional_info,omitempty"`
	TraceID        string          `json:"logging.googleapis.com/trace,omitempty"`
	Type           string          `json:"@type,omitempty"`
	SourceLocation *sourceLocation `json:"logging.googleapis.com/sourceLocation,omitempty"`
}

type sourceLocation struct {
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Function string `json:"function,omitempty"`
}

const errorType = "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent"

// Format logrus output per Google Cloud guidelines.
// See https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry for details.
// Tested with Google App Engine
//
// To have traceIDs logged, supply a key in the TraceKey field, and call logrus.WithContext(...)
//
// See the examples for usage.
func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	e := googleLogEntry{
		Message:    entry.Message,
		Severity:   entry.Level.String(),
		Additional: entry.Data,
	}

	if entry.Caller != nil {
		e.SourceLocation = &sourceLocation{
			File:     path.Base(entry.Caller.File),
			Line:     entry.Caller.Line,
			Function: entry.Caller.Function,
		}
	}

	if entry.Level == logrus.ErrorLevel {
		e.Type = errorType
		e.Message = entry.Message + "\n" + string(debug.Stack())
	}

	if entry.Context != nil {
		if traceid, ok := entry.Context.Value(f.TraceKey).(string); ok {
			e.TraceID = traceid
		}
	}

	serialized, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return append(serialized, '\n'), nil
}
