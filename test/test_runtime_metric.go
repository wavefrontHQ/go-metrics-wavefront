
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
		"runtime-metric": "1"
	}

	counter := metrics.NewCounter()                //Create a counter
	metrics.Register("golangruntime", counter)              // will create a 'some.prefix.golangruntime.count' metric with no tags
	reporting.RegisterMetric("golangruntime", counter, tags) // will create a 'some.prefix.golangruntime.count' metric with tags

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

	reporting.NewReporter(
		sender,
		application.New("app", "srv"),
		reporting.Source("go-metrics-test"),
		reporting.Prefix("some.prefix"),
		reporting.RuntimeMetric(true),
	)

	r.Start()

	fmt.Println("Search wavefront: ts(\"some.prefix.golangruntime.count\")")
	fmt.Println("Entering loop to simulate metrics flushing. Hit ctrl+c to cancel")

	for i:=0; i<=1000; i++ {
		time.Sleep(time.Second * 10)
	}
}
