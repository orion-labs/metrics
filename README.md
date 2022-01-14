# Prometheus Metrics Server Wrapper

This project provides a simple API used to hook into a prometheus metrics server.
Prometheus only requires that metrics are available at `/metrics` and  that the output
is in an expected format.

Formatting is done via the third party library [github.com/segmentio/stats/v4](https://github.com/segmentio/stats)

## Integration
### Setup
Add the packages to your project
```bash
go get github.com/orion-labs/metrics
```

### Startup
Integration with this API begins with setting up the server and stats engine. A default
is provided.
```go
package main
import (
     orion "github.com/orion-labs/metrics"
)

func main() {
    metrics := orion.NewDefaultMetrics()
    go metrics.Run()
}
```

More complex setup can be found in the [api/examples_test.go](./pkg/api/examples_test.go) file.

### Collection
Stats are tracked via the global engine initialized within the segmentio package. 
This means that stats can be collected without passing this metrics server arround. 

#### Tracking
```
package yourpackage
import (
	"bytes"
	"fmt"
	"github.com/segmentio/stats/v4"
	"github.com/segmentio/stats/v4/prometheus"
	"net/http/httptest"
	"time"
)

func somethingToDo() {
    stats.Incr("stat_one")
    stats.Add("stat_one", 10)
    stats.Incr("stat_too", stats.Tag{
        Name:  "a_tag",
    	Value: "custom_tag",
    })
}
```
These can be queried in prometheus via the following: `<prefix>_<statname>{a_tag="custom_tag"}`

