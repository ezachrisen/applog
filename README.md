# Applog

Applog formats logrus output for Google AppEngine:
- Errors are sent to Google Error Reporting with a stacktrace
- Code calling location is formatted with file, line and module
- Trace ID provided in the context is logged appropriately

```go 
import (
	"context"
	"os"

	"github.com/ezachrisen/applog"
	"github.com/sirupsen/logrus"
)

func main() {

	logrus.SetFormatter(&applog.Formatter{})
	logrus.Info("Hello")
}
```

See the example for additional usage. 

