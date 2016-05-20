
This is a plugin for [go-metrics](https://github.com/rcrowley/go-metrics) which adds a Wavefront reporter and a simple abstraction that supports tagging at the host and metric level.

### Usage

#### Wavefront Reporter

The Wavefront Reporter supports tagging at the host level. Any tags passed to the reporter here will be applied to every metric before being sent to Wavefront.

```go
import (
	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront"
)

func main() {
  hostTags := map[string]string{
    "source": "go-metrics-test",
  }
  go wavefront.Wavefront(metrics.DefaultRegistry, 1*time.Second, hostTags, "some.prefix", addr)
}
```
#### Tagging Metrics

In addition to tagging at the host level, you can add tags to individual metics.

```go
import (
	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront"
)

func main() {

	c := metrics.NewCounter()
	wavefront.RegisterMetric(
		"foo", c, map[string]string{
			"key1": "val1",
			"key2": "val2",
		})
	c.Inc(47)
}
```
`wavefront.RegisterMetric()` has the same affect as go-metrics' `metrics.Register()` except that it accepts tags in the form of a string map. The tags are then used by the Wavefront reporter at flush time. The tags becomes part of the key for a metric. Every unique combination of metric name+tags is a unique metric. You can pass your tags in any order to the Register and Get functions documented below. The Wavefront plugin ensures they are always encoded in the same order to ensure no duplication of metrics.

#### Extended Code Example

```go
package main

import (
	"fmt"
	"net"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront"
)

func main() {

	//Create a counter
	c := metrics.NewCounter()
	//Tags we'll add to the metric
	tags := map[string]string{
		"key2": "val1",
		"key1": "val2",
		"key0": "val0",
		"key4": "val4",
		"key3": "val3",
	}
	// Register it using wavefront.RegisterMetric instead of metrics.Register if there are tags
	wavefront.RegisterMetric("foo", c, tags)
	c.Inc(47)

	// Retreive it using our key and tags.
	// Any unique set of key+tags will be a unique series and thus a unique metric
	m2 := wavefront.GetMetric("foo", tags)
	fmt.Println(m2) // will print &{47}

	// Retrive it using wavefront.GetOrRegisterMetric instead of metrics.GetOrRegister if there are tags.
	m3 := wavefront.GetOrRegisterMetric("foo", c, tags)
	fmt.Println(m3) // will print &{47}

	//Let's remove the metric now
	wavefront.UnregisterMetric("foo", tags)

	//Try to get it after unregistering
	m4 := wavefront.GetMetric("foo", tags)
	fmt.Println(m4) // will print <nil>

	//Lets add it again and send it to Wavefront
	wavefront.RegisterMetric("foo", c, tags)
	c.Inc(47)

	// Set the address of the Wavefront Proxy
	addr, _ := net.ResolveTCPAddr("tcp", "192.168.99.100:2878")

	// Tags can be passed to the host as well (each tag will get applied to every metric)
	hostTags := map[string]string{
		"source": "go-metrics-test",
	}

	go wavefront.Wavefront(metrics.DefaultRegistry, 1*time.Second, hostTags, "some.prefix", addr)

	fmt.Println("Search wavefront: ts(\"some.prefix.foo.count\")")

	fmt.Println("Entering loop to simulate metrics flushing. Hit ctrl+c to cancel")
	for {
	}

}
```
### Go Docs


# wavefront
--
    import "github.com/wavefronthq/go-metrics-wavefront"


## Usage

#### func  DecodeKey

```go
func DecodeKey(key string) (string, string)
```
DecodeKey Decodes a metric key into a metric name and tag string

#### func  EncodeKey

```go
func EncodeKey(key string, tags map[string]string) string
```
EncodeKey Encodes the metric name and tags into a unique key.

#### func  GetMetric

```go
func GetMetric(key string, tags map[string]string) interface{}
```
GetMetric Tag support for metrics.Get()

#### func  GetOrRegisterMetric

```go
func GetOrRegisterMetric(name string, i interface{}, tags map[string]string) interface{}
```
GetOrRegisterMetric Tag support for metrics.GetOrRegister()

#### func  RegisterMetric

```go
func RegisterMetric(key string, metric interface{}, tags map[string]string)
```
RegisterMetric Tag support for metrics.Register()

#### func  UnregisterMetric

```go
func UnregisterMetric(name string, tags map[string]string)
```
UnregisterMetric Tag support for metrics.UnregisterMetric()

#### func  Wavefront

```go
func Wavefront(r metrics.Registry, d time.Duration, ht map[string]string, prefix string, addr *net.TCPAddr)
```
Wavefront is an exporter function which reports metrics to a wavefront proxy
located at addr, flushing them every d duration.

#### func  WavefrontOnce

```go
func WavefrontOnce(c WavefrontConfig) error
```
WavefrontOnce performs a single submission to Wavefront, returning a non-nil
error on failed connections. This can be used in a loop similar to
WavefrontWithConfig for custom error handling.

#### func  WavefrontWithConfig

```go
func WavefrontWithConfig(c WavefrontConfig)
```
WavefrontWithConfig calls Wavefront() but allows you to pass a WavefrontConfig
struct

#### type WavefrontConfig

```go
type WavefrontConfig struct {
	Addr          *net.TCPAddr     // Network address to connect to
	Registry      metrics.Registry // Registry to be exported
	FlushInterval time.Duration    // Flush interval
	DurationUnit  time.Duration    // Time conversion unit for durations
	Prefix        string           // Prefix to be prepended to metric names
	Percentiles   []float64        // Percentiles to export from timers and histograms
	HostTags      map[string]string
}
```

WavefrontConfig provides configuration parameters for the Wavefront exporter
