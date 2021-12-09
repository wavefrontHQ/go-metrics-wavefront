package reporting

import (
	"math"
	"testing"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
)

var compression = histogram.Compression(5)
var pow10, inc100, inc1000, emptyHistogram = NewHistogram(compression), NewHistogram(compression), NewHistogram(compression), NewHistogram(compression)

func createPow10Histogram() {
	pow10_ := [9]int64{0, 1, 10, 10, 100, 1000, 10000, 10000, 100000}
	for _, num := range pow10_ {
		pow10.Update(num)
	}
}

func setup() {
	createPow10Histogram()
	for i := 1; i <= 100; i++ {
		inc100.Update(int64(i))
	}
	for i := 1; i <= 1000; i++ {
		inc1000.Update(int64(i))
	}
	time.Sleep(1 * time.Minute) //flush to priorTimedBin
}

func TestHistogramCal(t *testing.T) {
	setup()
	//Count
	assert.Equal(t, int64(9), pow10.Count())
	assert.Equal(t, int64(9), pow10.Snapshot().Count())
	assert.Equal(t, int64(0), emptyHistogram.Count())
	//Max
	assert.Equal(t, int64(100000), pow10.Max())
	assert.Equal(t, int64(100000), pow10.Snapshot().Max())
	assert.Equal(t, int64(math.NaN()), emptyHistogram.Max())
	//Min
	assert.Equal(t, int64(1), inc100.Min())
	assert.Equal(t, int64(1), inc100.Snapshot().Min())
	assert.Equal(t, int64(math.NaN()), emptyHistogram.Min())
	//Mean
	assert.Equal(t, float64(13457.888888888889), pow10.Mean())
	assert.Equal(t, float64(13457.888888888889), pow10.Snapshot().Mean())
	assert.True(t, math.IsNaN(emptyHistogram.Mean()))
	//Sum
	assert.Equal(t, int64(121121), pow10.Sum())
	assert.Equal(t, int64(121121), pow10.Snapshot().Sum())
	assert.Equal(t, int64(0), emptyHistogram.Sum())
	//StdDev
	stddev := pow10.StdDev()
	assert.Equal(t, float64(30859.857493890177), stddev)
	assert.Equal(t, float64(0), emptyHistogram.StdDev())
	//Percentiles
	snapshot := inc100.Snapshot()
	assert.Equal(t, 25.25, snapshot.Percentile(0.25))
	assert.Equal(t, 75.75, snapshot.Percentile(0.75))
	assert.Equal(t, 98.98, snapshot.Percentile(0.98))
	assert.Equal(t, 99.99, snapshot.Percentile(0.99))
	assert.Equal(t, 999.999, inc1000.Snapshot().Percentile(0.999))
}

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
