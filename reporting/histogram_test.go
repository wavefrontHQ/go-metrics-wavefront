package reporting

import (
	"testing"

	metrics "github.com/rcrowley/go-metrics"
)

func TestWFHistogramAPI(t *testing.T) {
	h := NewHistogram()

	switch h.(type) {
	case metrics.Histogram:
		t.Log("-- metrics.Histogram --")
	default:
		t.Fatalf("the histogram is not 'metrics.Histogram'")
	}

	switch h.(type) {
	case Histogram:
		t.Log("-- reporter.Histogram --")
	default:
		t.Fatalf("the histogram is not 'histogram.Histogram'")
	}
}
