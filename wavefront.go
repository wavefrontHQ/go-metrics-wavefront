// temp

// Package wavefront is a plugin for go-metrics that provides a Wavefront reporter and tag support at the host and metric level.
package wavefront

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rcrowley/go-metrics"
)

// RegisterMetric tag support for metrics.Register()
func RegisterMetric(key string, metric interface{}, tags map[string]string) {
	key = EncodeKey(key, tags)
	metrics.Register(key, metric)
}

// GetMetric tag support for metrics.Get()
func GetMetric(key string, tags map[string]string) interface{} {
	key = EncodeKey(key, tags)
	return metrics.Get(key)
}

// GetOrRegisterMetric tag support for metrics.GetOrRegister()
func GetOrRegisterMetric(name string, i interface{}, tags map[string]string) interface{} {
	key := EncodeKey(name, tags)
	return metrics.GetOrRegister(key, i)
}

// UnregisterMetric tag support for metrics.UnregisterMetric()
func UnregisterMetric(name string, tags map[string]string) {
	key := EncodeKey(name, tags)
	metrics.Unregister(key)
}

// EncodeKey encodes the metric name and tags into a unique key.
func EncodeKey(key string, tags map[string]string) string {
	//sort the tags to ensure the key is always the same when getting or setting
	sortedKeys := make([]string, len(tags))
	i := 0
	for k, _ := range tags {
		sortedKeys[i] = k
		i++
	}
	sort.Strings(sortedKeys)
	keyAppend := "["
	for i := range sortedKeys {
		keyAppend += " " + sortedKeys[i] + "=\"" + tags[sortedKeys[i]] + "\""
	}
	keyAppend += "]"
	key += keyAppend
	return key
}

// DecodeKey decodes a metric key into a metric name and tag string
func DecodeKey(key string) (string, string) {
	if strings.Contains(key, "[") == false {
		return key, ""
	}
	parts := strings.Split(key, "[")
	name := parts[0]
	tagStr := parts[1]
	tagStr = tagStr[0 : len(tagStr)-1]
	return name, tagStr
}

func hostTagString(hostTags map[string]string) string {
	htStr := ""
	for k, v := range hostTags {
		htStr += " " + k + "=\"" + v + "\""
	}
	return htStr
}

// WavefrontConfig provides configuration parameters for
// the Wavefront exporter
type WavefrontConfig struct {
	Addr          *net.TCPAddr     // Network address to connect to
	Registry      metrics.Registry // Registry to be exported
	FlushInterval time.Duration    // Flush interval
	DurationUnit  time.Duration    // Time conversion unit for durations
	Prefix        string           // Prefix to be prepended to metric names
	Percentiles   []float64        // Percentiles to export from timers and histograms
	HostTags      map[string]string
}

// Wavefront is an exporter function which reports metrics to a
// wavefront proxy located at addr, flushing them every d duration.
func Wavefront(r metrics.Registry, d time.Duration, ht map[string]string, prefix string, addr *net.TCPAddr) {
	WavefrontWithConfig(WavefrontConfig{
		Addr:          addr,
		Registry:      r,
		FlushInterval: d,
		DurationUnit:  time.Nanosecond,
		Prefix:        prefix,
		HostTags:      ht,
		Percentiles:   []float64{0.5, 0.75, 0.95, 0.99, 0.999},
	})
}

// WavefrontWithConfig calls Wavefront() but allows you to pass a WavefrontConfig struct
func WavefrontWithConfig(c WavefrontConfig) {
	for _ = range time.Tick(c.FlushInterval) {
		if err := writeEntireRegistryAndFlush(&c); nil != err {
			log.Println(err)
		}
	}
}

// WavefrontOnce performs a single submission to Wavefront, returning a
// non-nil error on failed connections. This can be used in a loop
// similar to WavefrontWithConfig for custom error handling.
func WavefrontOnce(c WavefrontConfig) error {
	return writeEntireRegistryAndFlush(&c)
}

// WavefrontSingleMetric submits a single metric to Wavefront. The given metric
// is not registered in the underyling `go-metrics` registry and the registry
// will not be flushed entirely (unlike `WavefrontOnce`). If the connection to
// the Wavefront proxy cannot be made, a non-nil error is returned.
func WavefrontSingleMetric(c *WavefrontConfig, name string, metric interface{}, tags map[string]string) error {
	now := time.Now().Unix()
	conn, err := net.DialTCP("tcp", nil, c.Addr)
	if nil != err {
		return err
	}
	defer conn.Close()
	w := bufio.NewWriter(conn)

	key := EncodeKey(name, tags)
	WriteMetricAndFlush(w, metric, key, now, c)
	return nil
}

func writeEntireRegistryAndFlush(c *WavefrontConfig) error {
	now := time.Now().Unix()
	conn, err := net.DialTCP("tcp", nil, c.Addr)
	if nil != err {
		return err
	}
	defer conn.Close()
	w := bufio.NewWriter(conn)

	c.Registry.Each(func(key string, metric interface{}) {
		WriteMetricAndFlush(w, metric, key, now, c)
	})
	return nil
}

func WriteMetricAndFlush(w *bufio.Writer, i interface{}, key string, ts int64, c *WavefrontConfig) {
	name, tagStr := DecodeKey(key)
	tagStr += hostTagString(c.HostTags)

	switch metric := i.(type) {
	case metrics.Counter:
		writeCounter(w, metric, name, tagStr, ts, c)
	case metrics.Gauge:
		writeGauge(w, metric, name, tagStr, ts, c)
	case metrics.GaugeFloat64:
		writeGaugeFloat64(w, metric, name, tagStr, ts, c)
	case metrics.Histogram:
		writeHistogram(w, metric, name, tagStr, ts, c)
	case metrics.Meter:
		writeMeter(w, metric, name, tagStr, ts, c)
	case metrics.Timer:
		writeTimer(w, metric, name, tagStr, ts, c)
	}
	w.Flush()
}

func writeCounter(w *bufio.Writer, metric metrics.Counter, name, tagStr string, ts int64, c *WavefrontConfig) {
	fmt.Fprintf(w, "%s.%s.count %d %d %s\n", c.Prefix, name, metric.Count(), ts, tagStr)
}

func writeGauge(w *bufio.Writer, metric metrics.Gauge, name, tagStr string, ts int64, c *WavefrontConfig) {
	fmt.Fprintf(w, "%s.%s.value %d %d %s\n", c.Prefix, name, metric.Value(), ts, tagStr)
}

func writeGaugeFloat64(w *bufio.Writer, metric metrics.GaugeFloat64, name, tagStr string, ts int64, c *WavefrontConfig) {
	fmt.Fprintf(w, "%s.%s.value %f %d %s\n", c.Prefix, name, metric.Value(), ts, tagStr)
}

func writeHistogram(w *bufio.Writer, metric metrics.Histogram, name, tagStr string, ts int64, c *WavefrontConfig) {
	h := metric.Snapshot()
	ps := h.Percentiles(c.Percentiles)
	fmt.Fprintf(w, "%s.%s.count %d %d %s\n", c.Prefix, name, h.Count(), ts, tagStr)
	fmt.Fprintf(w, "%s.%s.min %d %d %s\n", c.Prefix, name, h.Min(), ts, tagStr)
	fmt.Fprintf(w, "%s.%s.max %d %d %s\n", c.Prefix, name, h.Max(), ts, tagStr)
	fmt.Fprintf(w, "%s.%s.mean %.2f %d %s\n", c.Prefix, name, h.Mean(), ts, tagStr)
	fmt.Fprintf(w, "%s.%s.std-dev %.2f %d %s\n", c.Prefix, name, h.StdDev(), ts, tagStr)
	for psIdx, psKey := range c.Percentiles {
		key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
		fmt.Fprintf(w, "%s.%s.%s-percentile %.2f %d %s\n", c.Prefix, name, key, ps[psIdx], ts, tagStr)
	}
}

func writeMeter(w *bufio.Writer, metric metrics.Meter, name, tagStr string, ts int64, c *WavefrontConfig) {
	m := metric.Snapshot()
	fmt.Fprintf(w, "%s.%s.count %d %d %s\n", c.Prefix, name, m.Count(), ts, tagStr)
	fmt.Fprintf(w, "%s.%s.one-minute %.2f %d %s\n", c.Prefix, name, m.Rate1(), ts, tagStr)
	fmt.Fprintf(w, "%s.%s.five-minute %.2f %d %s\n", c.Prefix, name, m.Rate5(), ts, tagStr)
	fmt.Fprintf(w, "%s.%s.fifteen-minute %.2f %d %s\n", c.Prefix, name, m.Rate15(), ts, tagStr)
	fmt.Fprintf(w, "%s.%s.mean %.2f %d %s\n", c.Prefix, name, m.RateMean(), ts, tagStr)
}

func writeTimer(w *bufio.Writer, metric metrics.Timer, name, tagStr string, ts int64, c *WavefrontConfig) {
	t := metric.Snapshot()
	du := float64(c.DurationUnit)
	ps := t.Percentiles(c.Percentiles)
	fmt.Fprintf(w, "%s.%s.count %d %d %s\n", c.Prefix, name, t.Count(), ts, tagStr)
	fmt.Fprintf(w, "%s.%s.min %d %d %s\n", c.Prefix, name, t.Min()/int64(du), ts, tagStr)
	fmt.Fprintf(w, "%s.%s.max %d %d %s\n", c.Prefix, name, t.Max()/int64(du), ts, tagStr)
	fmt.Fprintf(w, "%s.%s.mean %.2f %d %s\n", c.Prefix, name, t.Mean()/du, ts, tagStr)
	fmt.Fprintf(w, "%s.%s.std-dev %.2f %d %s\n", c.Prefix, name, t.StdDev()/du, ts, tagStr)
	for psIdx, psKey := range c.Percentiles {
		key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
		fmt.Fprintf(w, "%s.%s.%s-percentile %.2f %d %s\n", c.Prefix, name, key, ps[psIdx]/du, ts, tagStr)
	}
	fmt.Fprintf(w, "%s.%s.one-minute %.2f %d %s\n", c.Prefix, name, t.Rate1(), ts, tagStr)
	fmt.Fprintf(w, "%s.%s.five-minute %.2f %d %s\n", c.Prefix, name, t.Rate5(), ts, tagStr)
	fmt.Fprintf(w, "%s.%s.fifteen-minute %.2f %d %s\n", c.Prefix, name, t.Rate15(), ts, tagStr)
	fmt.Fprintf(w, "%s.%s.mean-rate %.2f %d %s\n", c.Prefix, name, t.RateMean(), ts, tagStr)
}
