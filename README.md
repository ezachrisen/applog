# Applog

Applog formats [logrus](https://github.com/sirupsen/logrus) output for Google AppEngine:
- Errors are sent to Google Error Reporting with a stacktrace
- Code calling location is formatted with file, line and module
- Trace ID provided in the context is logged appropriately

### Basic Usage

```go 
import (
	"github.com/ezachrisen/applog"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&applog.Formatter{})
	logrus.Info("Hello")
	// Output: {"message":"Hello","severity":"info"}
}
```

### TraceID in Context

```go

func main() {
	type ContextKey string
	key := ContextKey("traceid")

	logrus.SetFormatter(&applog.Formatter{TraceKey: key})
	ctx := context.WithValue(context.Background(), key, "abcdef123456")
	logrus.WithContext(ctx).Infof("My info here %d", 100)

	// Output:
	// {"message":"My info here 100","severity":"info","logging.googleapis.com/trace":"abcdef123456"}
}
```

### Benchmarks
Most operations are very fast, even obtaining the source-code calling location. Note that getting a full stack trace is quite expensive. 

See the benchmark_test.go file for specifics. 

```
BenchmarkBasic-8             	 1463979	       809 ns/op
BenchmarkBasicWithCaller-8   	  325903	      3713 ns/op
BenchmarkStructured-8        	  489372	      2505 ns/op
BenchmarkError-8             	   49635	     23598 ns/op
```
