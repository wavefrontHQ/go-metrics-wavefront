package main

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront"
)

func main() {

	c := metrics.NewCounter()
	wavefront.RegisterMetric(
		"foo", c, map[string]string{
			"key2": "val1",
			"key1": "val2",
			"key0": "val0",
			"key4": "val4",
			"key3": "val3",
		})

	c.Inc(47)

	g := metrics.NewGauge()
	metrics.Register("bar", g)
	g.Update(47)

	s := metrics.NewExpDecaySample(1028, 0.015) // or metrics.NewUniformSample(1028)
	h := metrics.NewHistogram(s)
	metrics.Register("baz", h)
	h.Update(47)

	m := metrics.NewMeter()
	metrics.Register("quux", m)
	m.Mark(47)

	t := metrics.NewTimer()
	metrics.Register("bang", t)
	t.Time(func() {})
	t.Update(47)

	addr, _ := net.ResolveTCPAddr("tcp", "192.168.99.100:2878")

	hostTags := map[string]string{
		"source": "go-metrics-test",
	}
	go wavefront.Wavefront(metrics.DefaultRegistry, 1*time.Second, hostTags, "some.prefix", addr)

	go metrics.Log(metrics.DefaultRegistry, 5*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))

	for {

	}

}
