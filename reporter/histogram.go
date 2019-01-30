package reporter

import (
	metrics "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
)

type Histogram struct {
	delgate histogram.Histogram
}

func NewHistogram(options ...histogram.Option) metrics.Histogram {
	return Histogram{delgate: histogram.New(options...)}
}

// Clear all samples on the histogram
func (h Histogram) Clear() {
	panic("Clear called on a HistogramSnapshot")
}

func (h Histogram) Count() int64 {
	return int64(h.delgate.Count())
}

func (h Histogram) Min() int64 {
	return int64(h.delgate.Min())
}

func (h Histogram) Max() int64 {
	return int64(h.delgate.Max())
}

func (h Histogram) Sum() int64 {
	return int64(h.delgate.Sum())
}

func (h Histogram) Mean() float64 {
	return h.delgate.Mean()
}

func (h Histogram) Update(v int64) {
	h.delgate.Update(float64(v))
}

// Sample will panic
func (h Histogram) Sample() metrics.Sample {
	panic("Sample called on a HistogramSnapshot")
}

// Snapshot will panic
// no need for a snapshot
func (h Histogram) Snapshot() metrics.Histogram {
	panic("Snapshot called on a HistogramSnapshot")
}

// TODO: review Min, Max, etc....
func (h Histogram) StdDev() float64 {
	panic("StdDev called on a HistogramSnapshot")
}

func (h Histogram) Variance() float64 {
	panic("Variance called on a HistogramSnapshot")
}

// Percentile returns the desired percentile estimation.
func (h Histogram) Percentile(p float64) float64 {
	return h.delgate.Quantile(p)
}

func (h Histogram) Percentiles(ps []float64) []float64 {
	var res []float64
	for _, p := range ps {
		res = append(res, h.Percentile(p))
	}
	return res
}

func (h Histogram) Distributions() []histogram.Distribution {
	return h.delgate.Distributions()
}

func (h Histogram) Granularity() histogram.HistogramGranularity {
	return h.delgate.Granularity()
}
