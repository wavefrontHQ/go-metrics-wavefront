package reporter

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/wavefront-sdk-go/application"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/senders"

	"github.com/stretchr/testify/assert"
)

func TestPrefixAndSuffix(t *testing.T) {
	reporter := &reporter{}

	reporter.prefix = "prefix"
	reporter.addSuffix = true
	name := reporter.prepareName("name", "count")
	assert.Equal(t, name, "prefix.name.count")

	name = reporter.prepareName("name")
	assert.Equal(t, name, "prefix.name")

	reporter.prefix = ""
	reporter.addSuffix = false
	name = reporter.prepareName("name", "count")
	assert.Equal(t, name, "name")
}

func TestError(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()

	sender := &MockSender{}
	reporter := New(sender, application.New("app", "srv"))
	tags := map[string]string{"tag1": "tag"}

	RegisterMetric("", metrics.NewCounter(), tags)

	c := metrics.NewCounter()
	RegisterMetric("m1", c, tags)
	c.Inc(1)

	reporter.Report()
	time.Sleep(time.Millisecond) // wait for the error channel to work

	assert.Equal(t, 1, len(sender.Metrics))
	assert.Equal(t, int64(1), reporter.ErrorsCount())
}

func TestBasicCounter(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()

	sender := &MockSender{}
	reporter := New(sender, application.New("app", "srv"))
	tags := map[string]string{"tag1": "tag"}

	name := "counter"
	c := GetMetric(name, tags)
	if c == nil {
		c = metrics.NewCounter()
		RegisterMetric(name, c, tags)
	}
	c.(metrics.Counter).Inc(1)

	reporter.Report()

	assert.Equal(t, 1, len(sender.Metrics))
}

func TestWFHistogram(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()

	sender := &MockSender{}
	reporter := New(sender, application.New("app", "srv"))
	tags := map[string]string{"tag1": "tag"}

	h := NewHistogram(histogram.Granularity(histogram.SECOND))
	// h := NewHistogram()
	RegisterMetric("wf.histogram", h, tags)

	for i := 0; i < 1000; i++ {
		h.Update(rand.Int63())
	}

	time.Sleep(time.Second * 2) // wait until the histogram rotates

	reporter.Report()

	fmt.Printf("-> dis: %v\n", sender.Distributions)
	fmt.Printf("-> met: %v\n", sender.Metrics)

	assert.Equal(t, 1, len(sender.Distributions))
	assert.Equal(t, 0, len(sender.Metrics))
}

func TestHistogram(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()

	sender := &MockSender{}
	reporter := New(sender, application.New("app", "srv"))
	tags := map[string]string{"tag1": "tag"}

	s := metrics.NewExpDecaySample(1028, 0.015) // or metrics.NewUniformSample(1028)
	h := metrics.NewHistogram(s)
	RegisterMetric("mt.histogram", h, tags)

	for i := 0; i < 1000; i++ {
		h.Update(rand.Int63())
	}

	reporter.Report()

	fmt.Printf("-> dis: %v\n", sender.Distributions)
	fmt.Printf("-> met: %v\n", sender.Metrics)

	assert.Equal(t, 0, len(sender.Distributions))
	assert.Equal(t, 10, len(sender.Metrics))
}

func TestDeltaPoint(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()

	sender := &MockSender{}
	reporter := New(sender, application.New("app", "srv"))
	tags := map[string]string{"tag1": "tag"}

	counter := metrics.NewCounter()
	RegisterMetric(DeltaCounterName("foo"), counter, tags)

	counter.Inc(10)
	reporter.Report()
	fmt.Printf("-> Deltas: %v\n", sender.Deltas)
	assert.Equal(t, 1, len(sender.Deltas))

	counter.Inc(10)
	reporter.Report()

	fmt.Printf("-> Deltas: %v\n", sender.Deltas)
	fmt.Printf("-> Metrics: %v\n", sender.Metrics)
	assert.Equal(t, 2, len(sender.Deltas))
	assert.Equal(t, 0, len(sender.Metrics))
}

type MockSender struct {
	Distributions []string
	Metrics       []string
	Deltas        []string
}

func (s MockSender) Close() {}

func (s MockSender) SendEvent(name string, startMillis, endMillis int64, source string, tags map[string]string) error {
	return nil
}

func (s MockSender) SendSpan(name string, startMillis, durationMillis int64, source, traceID, spanID string, parents, followsFrom []string, tags []senders.SpanTag, spanLogs []senders.SpanLog) error {
	return nil
}

func (s *MockSender) SendDistribution(name string, centroids []histogram.Centroid, hgs map[histogram.HistogramGranularity]bool, ts int64, source string, tags map[string]string) error {
	s.Distributions = append(s.Distributions, name)
	return nil
}

func (s *MockSender) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	s.Deltas = append(s.Deltas, name)
	return nil
}

func (s *MockSender) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	if name == ".count" {
		return fmt.Errorf("empty metric name")
	}
	s.Metrics = append(s.Metrics, name)
	return nil
}

func (s MockSender) Flush() error {
	return nil
}

func (s MockSender) GetFailureCount() int64 {
	return 0
}

func (s MockSender) Start() {}
