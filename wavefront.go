// temp

package wavefront

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/rcrowley/go-metrics"
)

// RegisterMetric A simple wrapper around go-metrics metrics.Register function
func RegisterMetric(key string, metric interface{}, tags map[string]string) {
	keyAppend := "["
	for k, v := range tags {
		keyAppend += " " + k + "=\"" + v + "\""
	}
	keyAppend += "]"
	key += keyAppend
	metrics.Register(key, metric)
}

// DecodeKey return the metric name and the tag string from a metric key
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

func HostTagString(hostTags map[string]string) string {
	htStr := ""
	for k, v := range hostTags {
		htStr += " " + k + "=\"" + v + "\""
	}
	return htStr
}

// WavefrontConfig provides a container with configuration parameters for
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

// Wavefront is a blocking exporter function which reports metrics in r
// to a wavefront server located at addr, flushing them every d duration
// and prepending metric names with prefix.
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

// WavefrontWithConfig is a blocking exporter function just like Wavefront,
// but it takes a WavefrontConfig instead.
func WavefrontWithConfig(c WavefrontConfig) {
	for _ = range time.Tick(c.FlushInterval) {
		if err := wavefront(&c); nil != err {
			log.Println(err)
		}
	}
}

// WavefrontOnce performs a single submission to Wavefront, returning a
// non-nil error on failed connections. This can be used in a loop
// similar to WavefrontWithConfig for custom error handling.
func WavefrontOnce(c WavefrontConfig) error {
	return wavefront(&c)
}

func wavefront(c *WavefrontConfig) error {
	now := time.Now().Unix()
	du := float64(c.DurationUnit)
	conn, err := net.DialTCP("tcp", nil, c.Addr)
	if nil != err {
		return err
	}
	defer conn.Close()
	w := bufio.NewWriter(conn)
	c.Registry.Each(func(name string, i interface{}) {
		name, tagStr := DecodeKey(name)
		tagStr += HostTagString(c.HostTags)
		switch metric := i.(type) {
		case metrics.Counter:
			fmt.Fprintf(w, "%s.%s.count %d %d %s\n", c.Prefix, name, metric.Count(), now, tagStr)
		case metrics.Gauge:
			fmt.Fprintf(w, "%s.%s.value %d %d %s\n", c.Prefix, name, metric.Value(), now, tagStr)
		case metrics.GaugeFloat64:
			fmt.Fprintf(w, "%s.%s.value %f %d %s\n", c.Prefix, name, metric.Value(), now, tagStr)
		case metrics.Histogram:
			h := metric.Snapshot()
			ps := h.Percentiles(c.Percentiles)
			fmt.Fprintf(w, "%s.%s.count %d %d %s\n", c.Prefix, name, h.Count(), now, tagStr)
			fmt.Fprintf(w, "%s.%s.min %d %d %s\n", c.Prefix, name, h.Min(), now, tagStr)
			fmt.Fprintf(w, "%s.%s.max %d %d %s\n", c.Prefix, name, h.Max(), now, tagStr)
			fmt.Fprintf(w, "%s.%s.mean %.2f %d %s\n", c.Prefix, name, h.Mean(), now, tagStr)
			fmt.Fprintf(w, "%s.%s.std-dev %.2f %d %s\n", c.Prefix, name, h.StdDev(), now, tagStr)
			for psIdx, psKey := range c.Percentiles {
				key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
				fmt.Fprintf(w, "%s.%s.%s-percentile %.2f %d %s\n", c.Prefix, name, key, ps[psIdx], now, tagStr)
			}
		case metrics.Meter:
			m := metric.Snapshot()
			fmt.Fprintf(w, "%s.%s.count %d %d %s\n", c.Prefix, name, m.Count(), now, tagStr)
			fmt.Fprintf(w, "%s.%s.one-minute %.2f %d %s\n", c.Prefix, name, m.Rate1(), now, tagStr)
			fmt.Fprintf(w, "%s.%s.five-minute %.2f %d %s\n", c.Prefix, name, m.Rate5(), now, tagStr)
			fmt.Fprintf(w, "%s.%s.fifteen-minute %.2f %d %s\n", c.Prefix, name, m.Rate15(), now, tagStr)
			fmt.Fprintf(w, "%s.%s.mean %.2f %d %s\n", c.Prefix, name, m.RateMean(), now, tagStr)
		case metrics.Timer:
			t := metric.Snapshot()
			ps := t.Percentiles(c.Percentiles)
			fmt.Fprintf(w, "%s.%s.count %d %d %s\n", c.Prefix, name, t.Count(), now, tagStr)
			fmt.Fprintf(w, "%s.%s.min %d %d %s\n", c.Prefix, name, t.Min()/int64(du), now, tagStr)
			fmt.Fprintf(w, "%s.%s.max %d %d %s\n", c.Prefix, name, t.Max()/int64(du), now, tagStr)
			fmt.Fprintf(w, "%s.%s.mean %.2f %d %s\n", c.Prefix, name, t.Mean()/du, now, tagStr)
			fmt.Fprintf(w, "%s.%s.std-dev %.2f %d %s\n", c.Prefix, name, t.StdDev()/du, now, tagStr)
			for psIdx, psKey := range c.Percentiles {
				key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
				fmt.Fprintf(w, "%s.%s.%s-percentile %.2f %d %s\n", c.Prefix, name, key, ps[psIdx]/du, now, tagStr)
			}
			fmt.Fprintf(w, "%s.%s.one-minute %.2f %d %s\n", c.Prefix, name, t.Rate1(), now, tagStr)
			fmt.Fprintf(w, "%s.%s.five-minute %.2f %d %s\n", c.Prefix, name, t.Rate5(), now, tagStr)
			fmt.Fprintf(w, "%s.%s.fifteen-minute %.2f %d %s\n", c.Prefix, name, t.Rate15(), now, tagStr)
			fmt.Fprintf(w, "%s.%s.mean-rate %.2f %d %s\n", c.Prefix, name, t.RateMean(), now, tagStr)
		}
		w.Flush()
	})
	return nil
}
