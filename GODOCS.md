# wavefront
--
    import "github.com/wavefronthq/go-metrics-wavefront"

Package wavefront is a plugin for go-metrics that provides a Wavefront reporter
and tag support at the host and metric level.

## Usage

#### func  DecodeKey

```go
func DecodeKey(key string) (string, string)
```
DecodeKey decodes a metric key into a metric name and tag string

#### func  EncodeKey

```go
func EncodeKey(key string, tags map[string]string) string
```
EncodeKey encodes the metric name and tags into a unique key.

#### func  GetMetric

```go
func GetMetric(key string, tags map[string]string) interface{}
```
GetMetric tag support for metrics.Get()

#### func  GetOrRegisterMetric

```go
func GetOrRegisterMetric(name string, i interface{}, tags map[string]string) interface{}
```
GetOrRegisterMetric tag support for metrics.GetOrRegister()

#### func  RegisterMetric

```go
func RegisterMetric(key string, metric interface{}, tags map[string]string)
```
RegisterMetric tag support for metrics.Register()

#### func  UnregisterMetric

```go
func UnregisterMetric(name string, tags map[string]string)
```
UnregisterMetric tag support for metrics.UnregisterMetric()

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
