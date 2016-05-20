
This is a plugin for [go-metrics](https://github.com/rcrowley/go-metrics) that adds a reporter for Wavefront and a simple abstraction for supporting tags.

### Usage

#### Wavefront Reporter

The Wavefront Reporter supports tagging at the host level. Any tags passed to the reporter here will be applied to every metric before being sent to Wavefront.

```go
import (
	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront"
)

hostTags := map[string]string{
  "source": "go-metrics-test",
}
go wavefront.Wavefront(metrics.DefaultRegistry, 1*time.Second, hostTags, "some.prefix", addr)

```
