package applog

import (
	"encoding/json"
	"fmt"
	"path"
	"runtime/debug"

	"go.opencensus.io/trace"

	"github.com/sirupsen/logrus"
)

const (
	// gcpTraceHeaderKey = "gcp-bin-trace"
	// grpcHeader        = "grpc-trace-bin"
	requestMethod = "requestMethod"
	requestUrl    = "requestUrl"
	latency       = "latency"
	grpcCode      = "grpcCode"
	grpcMessage   = "grpcMessage"
	grpcDetails   = "grpcDetails"
)

type GRPCFormatter struct {
	ProjectID string
}

type googleLogEntry struct {
	Message        string          `json:"message"`
	Severity       string          `json:"severity"`
	Additional     logrus.Fields   `json:"additional_info,omitempty"`
	TraceID        string          `json:"logging.googleapis.com/trace,omitempty"`
	Type           string          `json:"@type,omitempty"`
	SourceLocation *sourceLocation `json:"logging.googleapis.com/sourceLocation,omitempty"`
	HttpRequest    HttpRequest     `json:"httpRequest,omitempty"`
	GRPCStatus     GRPCStatus      `json:"grpc,omitempty"`
}

// Log this so Cloud Logging will parse it and display the information on the header line
type HttpRequest struct {
	RequestMethod string `json:"requestMethod,omitempty"`
	RequestUrl    string `json:"requestUrl,omitempty"`
	Latency       string `json:"latency,omitempty"`
}

type GRPCStatus struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Details string `json:"details,omitempty"`
}

// {
// 	"requestMethod": string,
// 	"requestUrl": string,
// 	"requestSize": string,
// 	"status": integer,
// 	"responseSize": string,
// 	"userAgent": string,
// 	"remoteIp": string,
// 	"serverIp": string,
// 	"referer": string,
// 	"latency": string,
// 	"cacheLookup": boolean,
// 	"cacheHit": boolean,
// 	"cacheValidatedWithOriginServer": boolean,
// 	"cacheFillBytes": string,
// 	"protocol": string
//   }
type sourceLocation struct {
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Function string `json:"function,omitempty"`
}

const errorType = "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent"

// Format logrus output per Google Cloud guidelines.
// See https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry for details.
//
// See the examples for usage.
func (f *GRPCFormatter) Format(entry *logrus.Entry) ([]byte, error) {

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
		span := trace.FromContext(entry.Context)
		if span != nil {
			e.TraceID = fmt.Sprintf("projects/%s/traces/%v", f.ProjectID, span.SpanContext().TraceID)
		}
	}

	if entry.Data != nil {
		//		fmt.Printf("entry.Data (%s) = %v\n", e.TraceID, entry.Data)
		if requestMethod, ok := entry.Data[requestMethod]; ok && requestMethod != "" {
			e.HttpRequest = HttpRequest{
				RequestMethod: fmt.Sprintf("%v", requestMethod),
				RequestUrl:    fmt.Sprintf("%v", entry.Data[requestUrl]),
				Latency:       fmt.Sprintf("%s", entry.Data[latency]),
			}
		}

		if code, ok := entry.Data[grpcCode]; ok && code != "" {
			e.GRPCStatus = GRPCStatus{
				Code:    fmt.Sprintf("%v", code),
				Message: fmt.Sprintf("%s", entry.Data[grpcMessage]),
				Details: fmt.Sprintf("%v", entry.Data[grpcDetails]),
			}
		}
		e.Additional = entry.Data
	}

	serialized, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return append(serialized, '\n'), nil
}
