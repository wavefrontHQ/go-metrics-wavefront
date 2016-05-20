package main

import (
	"fmt"
	"net"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront"
)

func main() {

	//Create a counter
	c := metrics.NewCounter()
	//Tags we'll add to the metric
	tags := map[string]string{
		"key2": "val1",
		"key1": "val2",
		"key0": "val0",
		"key4": "val4",
		"key3": "val3",
	}
	// Register it using wavefront.RegisterMetric instead of metrics.Register if there are tags
	wavefront.RegisterMetric("foo", c, tags)
	c.Inc(47)

	// Retreive it using our key and tags.
	// Any unique set of key+tags will be a unique series and thus a unique metric
	m2 := wavefront.GetMetric("foo", tags)
	fmt.Println(m2) // will print &{47}

	// Retrive it using wavefront.GetOrRegisterMetric instead of metrics.GetOrRegister if there are tags.
	m3 := wavefront.GetOrRegisterMetric("foo", c, tags)
	fmt.Println(m3) // will print &{47}

	//Let's remove the metric now
	wavefront.UnregisterMetric("foo", tags)

	//Try to get it after unregistering
	m4 := wavefront.GetMetric("foo", tags)
	fmt.Println(m4) // will print <nil>

	//Lets add it again and send it to Wavefront
	wavefront.RegisterMetric("foo", c, tags)
	c.Inc(47)

	// Set the address of the Wavefront Proxy
	addr, _ := net.ResolveTCPAddr("tcp", "192.168.99.100:2878")

	// Tags can be passed to the host as well (each tag will get applied to every metric)
	hostTags := map[string]string{
		"source": "go-metrics-test",
	}

	go wavefront.Wavefront(metrics.DefaultRegistry, 1*time.Second, hostTags, "some.prefix", addr)

	fmt.Println("Search wavefront: ts(\"some.prefix.foo.count\")")

	fmt.Println("Entering loop to simulate metrics flushing. Hit ctrl+c to cancel")
	for {
	}

}
