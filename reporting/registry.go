package reporting

import (
	"log"
	"reflect"

	metrics "github.com/rcrowley/go-metrics"
)

// RegisterMetric tag support for metrics.Register()
func RegisterMetric(name string, metric interface{}, tags map[string]string) {
	key := EncodeKey(name, tags)
	metrics.Register(key, metric)
	m := GetMetric(name, tags)
	if m == nil {
		log.Printf("[ERROR] metric '%s'(%s) not registered !!!", name, reflect.TypeOf(metric).String())
	}
}

// GetMetric tag support for metrics.Get()
func GetMetric(name string, tags map[string]string) interface{} {
	key := EncodeKey(name, tags)
	return metrics.Get(key)
}

// GetOrRegisterMetric tag support for metrics.GetOrRegister()
func GetOrRegisterMetric(name string, i interface{}, tags map[string]string) interface{} {
	key := EncodeKey(name, tags)
	return metrics.GetOrRegister(key, i)
}

// UnregisterMetric tag support for metrics.UnregisterMetric()
func UnregisterMetric(name string, tags map[string]string) {
	key := EncodeKey(name, tags)
	metrics.Unregister(key)
}
