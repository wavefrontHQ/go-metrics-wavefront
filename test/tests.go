package main

import (
	"fmt"
	"math/rand"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	wavefront "github.com/wavefronthq/go-metrics-wavefront/reporter"
	"github.com/wavefronthq/wavefront-sdk-go/application"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
)

func main() {

	//Tags we'll add to the metric
	tags := map[string]string{
		"key2": "val2",
		"key1": "val1",
		"key0": "val0",
		"key4": "val4",
		"key3": "val3",
	}

	counter := metrics.NewCounter()                //Create a counter
	metrics.Register("foo2", counter)              // will create a 'some.prefix.foo2.count' metric with no tags
	wavefront.RegisterMetric("foo", counter, tags) // will create a 'some.prefix.foo.count' metric with tags
	counter.Inc(47)

	histogram := histogram.New()
	metrics.Register("duration2", histogram)              // will do nothing
	wavefront.RegisterMetric("duration", histogram, tags) // will create a 'some.prefix.duration' histogram metric with tags

	deltaCounter := metrics.NewCounter()
	wavefront.RegisterMetric(wavefront.DeltaCounterName("delta.metric"), deltaCounter, tags)
	deltaCounter.Inc(10)

	directCfg := &senders.DirectConfiguration{
		Server:               "https://virunga.wavefront.com",
		Token:                "be4d70eb-c03a-4b93-b5d6-b8fd4b29629b",
		BatchSize:            10000,
		MaxBufferSize:        50000,
		FlushIntervalSeconds: 1,
	}

	sender, err := senders.NewDirectSender(directCfg)
	if err != nil {
		panic(err)
	}

	reporter := wavefront.New(
		sender,
		application.New("app", "srv"),
		wavefront.Source("go-metrics-test"),
		wavefront.Prefix("some.prefix"),
	)
	reporter.Start()

	fmt.Println("Search wavefront: ts(\"some.prefix.foo.count\")")
	fmt.Println("Entering loop to simulate metrics flushing. Hit ctrl+c to cancel")

	for {
		counter.Inc(rand.Int63())
		histogram.Update(rand.Int63())
		deltaCounter.Inc(10)
		time.Sleep(time.Second * 10)
	}
}
