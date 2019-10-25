package reporting

import (
	"testing"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
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

var pow10 = NewHistogram()

func setup() {
	pow10.Update(0)
	pow10.Update(1)
	pow10.Update(10)
	pow10.Update(10)
	pow10.Update(100)
	pow10.Update(1000)
	pow10.Update(10000)
	pow10.Update(10000)
	pow10.Update(100000)
}

func TestHistogram_StdDev(t *testing.T) {
	setup()
	time.Sleep(1 * time.Minute) //flush to priorTimedBin
	stddev := pow10.StdDev()
	assert.Equal(t, float64(30859.857493890177), stddev)
}
