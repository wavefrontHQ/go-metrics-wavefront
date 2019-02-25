package reporting

import (
	"fmt"
	"reflect"

	metrics "github.com/rcrowley/go-metrics"
)

// RegistryError returned if there is any error on RegisterMetric
type RegistryError string

func (err RegistryError) Error() string {
	return fmt.Sprintf("Registry Error: %s", string(err))
}

// RegisterMetric tag support for metrics.Register()
// return RegistryError if the metric is not registered
func RegisterMetric(name string, metric interface{}, tags map[string]string) error {
	key := EncodeKey(name, tags)
	err := metrics.Register(key, metric)
	if err != nil {
		return err
	}
	m := GetMetric(name, tags)
	if m == nil {
		return RegistryError(fmt.Sprintf("Metric '%s'(%s) not registered.", name, reflect.TypeOf(metric).String()))
	}
	return nil
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
