package reporting

import metrics "github.com/rcrowley/go-metrics"

// RegisterMetric tag support for metrics.Register()
func RegisterMetric(key string, metric interface{}, tags map[string]string) {
	key = EncodeKey(key, tags)
	metrics.Register(key, metric)
}

// GetMetric tag support for metrics.Get()
func GetMetric(key string, tags map[string]string) interface{} {
	key = EncodeKey(key, tags)
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
