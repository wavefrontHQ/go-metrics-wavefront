package reporting

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-sdk-go/application"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
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

func TestTags(t *testing.T) {
	sender := &MockSender{}
	reporter := NewReporter(sender, application.New("app", "srv"), DisableAutoStart(),
		LogErrors(true), CustomRegistry(metrics.NewRegistry()))

	reporter.GetOrRegisterMetric("m1", metrics.NewCounter(), map[string]string{"tag1": "tag"})
	reporter.GetOrRegisterMetric("m2", metrics.NewCounter(), map[string]string{"application": "tag"})

	reporter.Report()
	reporter.Close()

	assert.Equal(t, 2, len(sender.Metrics))
	for _, metric := range sender.Metrics {
		switch metric.Name {
		case "m1.count":
			assert.Equal(t, 5, len(metric.Tags))
			assert.Equal(t, "app", metric.Tags["application"], "metrics tags: %v", metric.Tags)
		case "m2.count":
			assert.Equal(t, 4, len(metric.Tags))
			assert.Equal(t, "tag", metric.Tags["application"], "metrics tags: %v", metric.Tags)
		default:
			t.Errorf("unexpected metric: '%v'", metric)
		}
	}
}

func TestError(t *testing.T) {
	sender := &MockSender{}
	reporter := NewReporter(sender, application.New("app", "srv"), DisableAutoStart(),
		LogErrors(true), CustomRegistry(metrics.NewRegistry()))
	tags := map[string]string{"tag1": "tag"}

	reporter.GetOrRegisterMetric("", metrics.NewCounter(), tags)

	c := metrics.NewCounter()
	reporter.RegisterMetric("m1", c, tags)
	c.Inc(1)

	reporter.Report()
	reporter.Close()
	time.Sleep(time.Second * 2)

	_, met, _ := sender.Counters()

	assert.Equal(t, 1, met)
	assert.NotEqual(t, int64(0), reporter.ErrorsCount(), "error count")
}

func TestBasicCounter(t *testing.T) {
	sender := &MockSender{}
	reporter := NewReporter(sender, application.New("app", "srv"), DisableAutoStart(),
		LogErrors(true), CustomRegistry(metrics.NewRegistry()))
	tags := map[string]string{"tag1": "tag"}

	name := "counter"
	c := reporter.GetMetric(name, tags)
	if c == nil {
		c = metrics.NewCounter()
		reporter.RegisterMetric(name, c, tags)
	}
	c.(metrics.Counter).Inc(1)

	for i := 0; i < 3; i++ {
		reporter.Report()
	}
	reporter.Close()

	_, met, _ := sender.Counters()
	assert.True(t, met >= 2)
}

func TestWFHistogram(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Histogram tests in short mode")
	}

	sender := newMockSender()
	reporter := NewReporter(sender, application.New("app", "srv"), DisableAutoStart(),
		LogErrors(true), CustomRegistry(metrics.NewRegistry()))
	tags := map[string]string{"tag1": "tag"}

	h := NewHistogram(histogram.GranularityOption(histogram.MINUTE))
	// h := NewHistogram()
	reporter.RegisterMetric("wf.histogram", h, tags)

	for i := 0; i < 1000; i++ {
		h.Update(rand.Int63())
	}

	time.Sleep(time.Minute * 2) // wait until the histogram rotates

	reporter.Report()

	dis, met, _ := sender.Counters()
	assert.Equal(t, 1, dis)
	assert.Equal(t, 0, met)

	reporter.Close()
}

func TestHistogram(t *testing.T) {
	sender := newMockSender()
	reporter := NewReporter(sender, application.New("app", "srv"), DisableAutoStart(),
		LogErrors(true), CustomRegistry(metrics.NewRegistry()))
	tags := map[string]string{"tag1": "tag"}

	s := metrics.NewExpDecaySample(1028, 0.015) // or metrics.NewUniformSample(1028)
	h := metrics.NewHistogram(s)
	reporter.RegisterMetric("mt.histogram", h, tags)

	for i := 0; i < 1000; i++ {
		h.Update(rand.Int63())
	}

	reporter.Report()
	dis, met, _ := sender.Counters()

	assert.Equal(t, 0, dis)
	assert.Equal(t, 10, met)

	reporter.Close()
}

func TestDeltaPoint(t *testing.T) {
	sender := newMockSender()
	reporter := NewReporter(sender, application.New("app", "srv"), DisableAutoStart(),
		LogErrors(true), CustomRegistry(metrics.NewRegistry()))
	tags := map[string]string{"tag1": "tag"}

	counter := metrics.NewCounter()
	reporter.RegisterMetric(DeltaCounterName("foo"), counter, tags)

	counter.Inc(10)
	reporter.Report()
	_, met, del := sender.Counters()
	assert.Equal(t, 1, del)
	assert.Equal(t, 0, met)

	counter.Inc(10)
	reporter.Report()

	_, met, del = sender.Counters()
	assert.Equal(t, 2, del)
	assert.Equal(t, 0, met)

	reporter.Close()
}

func newMockSender() *MockSender {
	return &MockSender{
		Distributions: make([]MockMetirc, 0),
		Metrics:       make([]MockMetirc, 0),
		Deltas:        make([]MockMetirc, 0),
	}
}

type MockMetirc struct {
	Name string
	Tags map[string]string
}

type MockSender struct {
	Distributions []MockMetirc
	Metrics       []MockMetirc
	Deltas        []MockMetirc
	sync.Mutex
}

func (s *MockSender) Close() {}

func (s *MockSender) SendEvent(name string, startMillis, endMillis int64, source string, tags map[string]string) error {
	return nil
}

func (s *MockSender) SendSpan(name string, startMillis, durationMillis int64, source, traceID, spanID string, parents, followsFrom []string, tags []senders.SpanTag, spanLogs []senders.SpanLog) error {
	return nil
}

func (s *MockSender) SendDistribution(name string, centroids []histogram.Centroid, hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error {
	s.Lock()
	defer s.Unlock()
	s.Distributions = append(s.Distributions, MockMetirc{Name: name, Tags: tags})
	return nil
}

func (s *MockSender) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	s.Lock()
	defer s.Unlock()
	s.Deltas = append(s.Deltas, MockMetirc{Name: name, Tags: tags})
	return nil
}

func (s *MockSender) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	if name == ".count" {
		return fmt.Errorf("empty metric name")
	}
	s.Lock()
	defer s.Unlock()
	s.Metrics = append(s.Metrics, MockMetirc{Name: name, Tags: tags})
	return nil
}

func (s *MockSender) Flush() error {
	return nil
}

func (s *MockSender) GetFailureCount() int64 {
	return 0
}

func (s *MockSender) Start() {}

func (s *MockSender) Counters() (int, int, int) {
	s.Lock()
	defer s.Unlock()
	return len(s.Distributions), len(s.Metrics), len(s.Deltas)
}
