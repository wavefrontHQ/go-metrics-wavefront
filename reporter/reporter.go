package reporter

import (
	"log"
	"strconv"
	"strings"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/wavefront-sdk-go/application"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	wf "github.com/wavefronthq/wavefront-sdk-go/senders"
)

// WavefrontMetricsReporter report go-metrics to wavefront
type WavefrontMetricsReporter interface {
	Start()
	Stop()
	Report()
	ErrorsCount() int64
}

type reporter struct {
	sender       wf.Sender
	application  application.Tags
	source       string
	prefix       string
	addSuffix    bool
	interval     time.Duration
	ticker       *time.Ticker
	percentiles  []float64              // Percentiles to export from timers and histograms
	durationUnit time.Duration          // Time conversion unit for durations
	metrics      map[string]interface{} // for Wavefron specific metrics tyoes, like Histograms
	errors       chan error
	errorsCount  int64
}

// Option allow WavefrontReporter customization
type Option func(*reporter)

// Source tag for metrics
func Source(source string) Option {
	return func(args *reporter) {
		args.source = source
	}
}

// ReportingInterval chanfe the default 1 minute reporting interval.
func ReportingInterval(interval time.Duration) Option {
	return func(args *reporter) {
		args.interval = interval
	}
}

// Prefix for the metrics name
func Prefix(prefix string) Option {
	return func(args *reporter) {
		args.prefix = strings.TrimSuffix(prefix, ".")
	}
}

// AddSuffix add a metric suffix based on the metric type ('.count', '.value')
func AddSuffix(addSuffix bool) Option {
	return func(args *reporter) {
		args.addSuffix = addSuffix
	}
}

// New create a WavefrontMetricsReporter
func New(sender wf.Sender, application application.Tags, setters ...Option) WavefrontMetricsReporter {
	r := &reporter{
		sender:       sender,
		application:  application,
		source:       "mySource",
		interval:     time.Second * 5,
		percentiles:  []float64{0.5, 0.75, 0.95, 0.99, 0.999},
		durationUnit: time.Nanosecond,
		metrics:      make(map[string]interface{}),
		addSuffix:    true,
		errorsCount:  0,
	}

	for _, setter := range setters {
		setter(r)
	}

	r.errors = make(chan error)
	go func() {
		for error := range r.errors {
			if error != nil {
				r.errorsCount++
				log.Printf("(%d) reporter error: %v\n", r.errorsCount, error)
			}
		}
	}()

	// r.Start()

	return r
}

func (r *reporter) ErrorsCount() int64 {
	return r.errorsCount
}

func (r *reporter) Report() {
	log.Printf("reporting....\n")
	metrics.DefaultRegistry.Each(func(key string, metric interface{}) {
		name, tags := DecodeKey(key)
		log.Printf("reporting metric: '%s'\n", r.prepareName(name))
		switch metric.(type) {
		case metrics.Counter:
			if hasDeltaPrefix(name) {
				r.reportDelta(name, metric.(metrics.Counter), tags)
			} else {
				r.errors <- r.sender.SendMetric(r.prepareName(name, "count"), float64(metric.(metrics.Counter).Count()), 0, r.source, tags)
			}
		case metrics.Gauge:
			r.errors <- r.sender.SendMetric(r.prepareName(name, "value"), float64(metric.(metrics.Gauge).Value()), 0, r.source, tags)
		case metrics.GaugeFloat64:
			r.errors <- r.sender.SendMetric(r.prepareName(name, "value"), float64(metric.(metrics.GaugeFloat64).Value()), 0, r.source, tags)
		case Histogram:
			r.reportWFHistogram(name, metric.(Histogram), tags)
		case metrics.Histogram:
			r.reportHistogram(name, metric.(metrics.Histogram), tags)
		case metrics.Meter:
			r.reportMeter(name, metric.(metrics.Meter), tags)
		case metrics.Timer:
			r.reportTimer(name, metric.(metrics.Timer), tags)
		}
	})
}

func (r *reporter) reportDelta(name string, metric metrics.Counter, tags map[string]string) {
	var prunedName string
	if strings.HasPrefix(name, deltaPrefix) {
		prunedName = name[deltaPrefixSize:]
	} else if strings.HasPrefix(name, altDeltaPrefix) {
		prunedName = name[altDeltaPrefixSize:]
	}
	value := metric.Count()
	metric.Dec(value)

	r.errors <- r.sender.SendDeltaCounter(deltaPrefix+r.prepareName(prunedName, "count"), float64(value), r.source, tags)
}

func (r *reporter) reportWFHistogram(metricName string, h Histogram, tags map[string]string) {
	distributions := h.Distributions()
	hgs := map[histogram.HistogramGranularity]bool{h.Granularity(): true}
	for _, distribution := range distributions {
		r.errors <- r.sender.SendDistribution(r.prepareName(metricName), distribution.Centroids, hgs, distribution.Timestamp.Unix(), r.source, tags)
	}
}

func (r *reporter) reportHistogram(name string, metric metrics.Histogram, tags map[string]string) {
	h := metric.Snapshot()
	ps := h.Percentiles(r.percentiles)
	r.errors <- r.sender.SendMetric(r.prepareName(name+".count"), float64(h.Count()), 0, r.source, tags)
	r.errors <- r.sender.SendMetric(r.prepareName(name+".min"), float64(h.Min()), 0, r.source, tags)
	r.errors <- r.sender.SendMetric(r.prepareName(name+".max"), float64(h.Max()), 0, r.source, tags)
	r.errors <- r.sender.SendMetric(r.prepareName(name+".mean"), h.Mean(), 0, r.source, tags)
	r.errors <- r.sender.SendMetric(r.prepareName(name+".std-dev"), h.StdDev(), 0, r.source, tags)
	for psIdx, psKey := range r.percentiles {
		key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
		r.errors <- r.sender.SendMetric(r.prepareName(name+"."+key+"-percentile"), ps[psIdx], 0, r.source, tags)
	}
}

func (r *reporter) reportMeter(name string, metric metrics.Meter, tags map[string]string) {
	m := metric.Snapshot()
	r.errors <- r.sender.SendMetric(r.prepareName(name+".count"), float64(m.Count()), 0, r.source, tags)
	r.errors <- r.sender.SendMetric(r.prepareName(name+".one-minute"), m.Rate1(), 0, r.source, tags)
	r.errors <- r.sender.SendMetric(r.prepareName(name+".five-minute"), m.Rate5(), 0, r.source, tags)
	r.errors <- r.sender.SendMetric(r.prepareName(name+".fifteen-minute"), m.Rate15(), 0, r.source, tags)
	r.errors <- r.sender.SendMetric(r.prepareName(name+".mean"), m.RateMean(), 0, r.source, tags)
}

func (r *reporter) reportTimer(name string, metric metrics.Timer, tags map[string]string) {
	t := metric.Snapshot()
	du := float64(r.durationUnit)
	ps := t.Percentiles(r.percentiles)
	r.errors <- r.sender.SendMetric(r.prepareName(name+".count"), float64(t.Count()), 0, r.source, tags)
	r.errors <- r.sender.SendMetric(r.prepareName(name+".min"), float64(t.Min()/int64(du)), 0, r.source, tags)
	r.errors <- r.sender.SendMetric(r.prepareName(name+".max"), float64(t.Max()/int64(du)), 0, r.source, tags)
	r.errors <- r.sender.SendMetric(r.prepareName(name+".mean"), t.Mean()/du, 0, r.source, tags)
	r.errors <- r.sender.SendMetric(r.prepareName(name+".std-dev"), t.StdDev()/du, 0, r.source, tags)
	for psIdx, psKey := range r.percentiles {
		key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
		r.errors <- r.sender.SendMetric(r.prepareName(name+"."+key+"-percentile"), ps[psIdx]/du, 0, r.source, tags)
	}
	r.errors <- r.sender.SendMetric(r.prepareName(name+".one-minute"), t.Rate1(), 0, r.source, tags)
	r.errors <- r.sender.SendMetric(r.prepareName(name+".five-minute"), t.Rate5(), 0, r.source, tags)
	r.errors <- r.sender.SendMetric(r.prepareName(name+".fifteen-minute"), t.Rate15(), 0, r.source, tags)
	r.errors <- r.sender.SendMetric(r.prepareName(name+".mean-rate"), t.RateMean(), 0, r.source, tags)
}

func (r *reporter) prepareName(name string, suffix ...string) string {
	if len(r.prefix) > 0 {
		name = r.prefix + "." + name
	}

	if r.addSuffix {
		for _, s := range suffix {
			name += "." + s
		}
	}

	return name
}

func (r *reporter) Start() {
	if r.ticker == nil {
		r.ticker = time.NewTicker(r.interval)
		go func() {
			for range r.ticker.C {
				r.Report()
			}
		}()
	}
}

func (r *reporter) Stop() {
	if r.ticker != nil {
		r.ticker.Stop()
		r.ticker = nil
	}
	r.Report()
}
