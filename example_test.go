package applog_test

import (
	"context"
	"os"

	"github.com/ezachrisen/applog"
	"github.com/sirupsen/logrus"
)

func ExampleBasic() {

	logrus.SetOutput(os.Stdout) // required for testing only
	logrus.SetFormatter(&applog.Formatter{})
	logrus.SetReportCaller(false)

	logrus.Info("Hello")

	logrus.WithFields(logrus.Fields{
		"animal": "walrus",
		"number": 1,
	}).Info("My info message here")

	// Output:
	// {"message":"Hello","severity":"info"}
	// {"message":"My info message here","severity":"info","additional_info":{"animal":"walrus","number":1}}

}

func ExampleContextTrace() {

	type ContextKey string
	key := ContextKey("traceid")

	logrus.SetOutput(os.Stdout) // required for testing only
	logrus.SetFormatter(&applog.Formatter{TraceKey: key})
	logrus.SetReportCaller(false)

	ctx := context.WithValue(context.Background(), key, "abcdef123456")
	logrus.WithContext(ctx).Infof("My info here %d", 100)

	// Output:
	// {"message":"My info here 100","severity":"info","logging.googleapis.com/trace":"abcdef123456"}
}
