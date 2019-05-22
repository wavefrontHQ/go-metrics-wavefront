# go-metrics-wavefront [![GoDoc](https://godoc.org/github.com/wavefrontHQ/go-metrics-wavefront?status.svg)](https://godoc.org/github.com/wavefrontHQ/go-metrics-wavefront) [![travis build status](https://travis-ci.com/wavefrontHQ/go-metrics-wavefront.svg?branch=master)](https://travis-ci.com/wavefrontHQ/go-metrics-wavefront)

This is a plugin for [go-metrics](https://github.com/rcrowley/go-metrics) which adds a Wavefront reporter and a simple abstraction that supports tagging at the host and metric level.

## Imports
```go
import (
	metrics "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"
	"github.com/wavefronthq/wavefront-sdk-go/application"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
)
```

## Set Up a Wavefront Reporter
This SDK provides a `WavefrontMetricsReporter` that allows you to:
* Report metrics to Wavefront at regular intervals or
* Manually report metrics to Wavefront

The steps for creating a `WavefrontMetricsReporter` are:
1. Create a Wavefront `Sender` for managing communication with Wavefront.
2. Create a `WavefrontMetricsReporter`

### 1. Set Up a Wavefront Sender
A "Wavefront sender" is an object that implements the low-level interface for sending data to Wavefront. You can choose to send data using either the [Wavefront proxy](https://docs.wavefront.com/proxies.html) or [direct ingestion](https://docs.wavefront.com/direct_ingestion.html).

* If you have already set up a Wavefront sender for another SDK that will run in the same process, use that one. (For details, see [Share a Wavefront Sender](https://github.com/wavefrontHQ/wavefront-sdk-go/blob/master/docs/sender.md#share-a-wavefront-sender).)
* Otherwise, follow the steps in [Set Up a Wavefront Sender](https://github.com/wavefrontHQ/wavefront-sdk-go/blob/master/docs/sender.md) to configure a proxy `Sender` or a direct `Sender`.

The following example configures a direct `Sender` with default direct ingestion properties:

```go
directCfg := &senders.DirectConfiguration{
  Server:               "https://INSTANCE.wavefront.com",
  Token:                "YOUR_API_TOKEN",
}

sender, err := senders.NewDirectSender(directCfg)
if err != nil {
  panic(err)
}
```

### 2. Create the WavefrontMetricsReporter
The `WavefrontMetricsReporter` supports tagging at the host level. Any tags passed to the reporter here will be applied to every metric before being sent to Wavefront.

To create the `WavefrontMetricsReporter` you initialize it with the `sender` instance you created in the previous step along with a few other properties:

```go
reporter := reporting.NewReporter(
  sender,
  application.New("app", "srv"),
  reporting.Source("go-metrics-test"),
  reporting.Prefix("some.prefix"),
  reporting.LogErrors(true),
)
```

## Tagging Metrics

In addition to tagging at the reporter level, you can add tags to individual metrics:

```go
tags := map[string]string{
  "key1": "val1",
  "key2": "val2",
}
counter := metrics.NewCounter() // Create a counter
reporter.RegisterMetric("foo", counter, tags) // will create a 'some.prefix.foo.count' metric with tags
counter.Inc(47)
```

## Extended Code Example

```go
package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"
	"github.com/wavefronthq/wavefront-sdk-go/application"
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

  // Create a direct sender
	directCfg := &senders.DirectConfiguration{
		Server:               "https://" + os.Getenv("WF_INSTANCE") + ".reporting.com",
		Token:                os.Getenv("WF_TOKEN"),
		BatchSize:            10000,
		MaxBufferSize:        50000,
		FlushIntervalSeconds: 1,
	}

	sender, err := senders.NewDirectSender(directCfg)
	if err != nil {
		panic(err)
	}

	reporter := reporting.NewReporter(
		sender,
		application.New("app", "srv"),
		reporting.Source("go-metrics-test"),
		reporting.Prefix("some.prefix"),
		reporting.LogErrors(true),
	)

	counter := metrics.NewCounter()                //Create a counter
	reporter.RegisterMetric("foo", counter, tags)  // will create a 'some.prefix.foo.count' metric with tags
	counter.Inc(47)

	histogram := reporting.NewHistogram()
	reporter.RegisterMetric("duration", histogram, tags) // will create a 'some.prefix.duration' histogram metric with tags

	histogram2 := reporting.NewHistogram()
	reporter.Register("duration2", histogram2) // will create a 'some.prefix.duration2' histogram metric with no tags

	deltaCounter := metrics.NewCounter()
	reporter.RegisterMetric(reporting.DeltaCounterName("delta.metric"), deltaCounter, tags)
	deltaCounter.Inc(10)

	fmt.Println("Search wavefront: ts(\"some.prefix.foo.count\")")
	fmt.Println("Entering loop to simulate metrics flushing. Hit ctrl+c to cancel")

	for {
		counter.Inc(rand.Int63())
		histogram.Update(rand.Int63())
		histogram2.Update(rand.Int63())
		deltaCounter.Inc(10)
		time.Sleep(time.Second * 10)
	}
}
```
