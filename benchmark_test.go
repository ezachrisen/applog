package applog

import (
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
)

func BenchmarkBasic(b *testing.B) {
	logrus.SetOutput(ioutil.Discard)
	logrus.SetFormatter(&Formatter{TraceKey: "ABC"})
	logrus.SetReportCaller(false)
	for i := 0; i < b.N; i++ {
		logrus.Info("My info message here")
	}
}

func BenchmarkBasicWithCaller(b *testing.B) {
	logrus.SetOutput(ioutil.Discard)
	logrus.SetFormatter(&Formatter{TraceKey: "ABC"})
	logrus.SetReportCaller(true)

	for i := 0; i < b.N; i++ {
		logrus.Info("My info message here")
	}
}

func BenchmarkStructured(b *testing.B) {
	logrus.SetOutput(ioutil.Discard)
	logrus.SetFormatter(&Formatter{TraceKey: "ABC"})
	logrus.SetReportCaller(false)

	for i := 0; i < b.N; i++ {
		logrus.WithFields(logrus.Fields{
			"animal": "walrus",
			"name":   "jonas",
			"age":    33,
		}).Info("My info message here")
	}
}

func BenchmarkError(b *testing.B) {
	logrus.SetOutput(ioutil.Discard)
	logrus.SetFormatter(&Formatter{TraceKey: "ABC"})

	for i := 0; i < b.N; i++ {
		logrus.Error("My error message here")
	}
}
