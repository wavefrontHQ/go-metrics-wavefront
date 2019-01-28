package reporter

import (
	"sync"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
)

type WavefrontRegistry struct {
	metrics.Registry
	metrics map[string]interface{}
	mutex   sync.RWMutex
}

func NewWavefrontRegistry() *WavefrontRegistry {
	return &WavefrontRegistry{
		Registry: metrics.DefaultRegistry,
		metrics:  make(map[string]interface{}),
	}
}

var DefaultWavefrontRegistry = NewWavefrontRegistry()

// Register the given metric under the given name.  Returns a DuplicateMetric
// if a metric by the given name is already registered.
func (r *WavefrontRegistry) Register(name string, i interface{}) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.register(name, i)
}

// Get the metric by the given name or nil if none is registered.
func (r *WavefrontRegistry) Get(name string) interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	metric, ok := r.metrics[name]
	if !ok {
		metric = r.Registry.Get(name)
	}
	return metric
}

// GetOrRegister Gets an existing metric or creates and registers a new one. Threadsafe
func (r *WavefrontRegistry) GetOrRegister(name string, i interface{}) interface{} {
	var res interface{}
	switch i.(type) {
	case histogram.Histogram:
		metric, ok := r.metrics[name]
		if !ok {
			r.register(name, i)
			res = i
		} else {
			res = metric
		}
	default:
		res = r.Registry.GetOrRegister(name, i)
	}
	return res
}

// Each call the given function for each registered metric.
func (r *WavefrontRegistry) Each(f func(string, interface{})) {
	for name, i := range r.registered() {
		f(name, i)
	}
	r.Registry.Each(f)
}

// UnregisterAll metrics.  (Mostly for testing.)
func (r *WavefrontRegistry) UnregisterAll() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for name := range r.metrics {
		delete(r.metrics, name)
	}
	r.Registry.UnregisterAll()
}

func (r *WavefrontRegistry) register(name string, i interface{}) error {
	if _, ok := r.metrics[name]; ok {
		return metrics.DuplicateMetric(name)
	}
	switch i.(type) {
	case histogram.Histogram:
		r.metrics[name] = i
	default:
		if error := r.Registry.Register(name, i); error != nil {
			return error
		}
	}
	return nil
}

func (r *WavefrontRegistry) registered() map[string]interface{} {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	metrics := make(map[string]interface{}, len(r.metrics))
	for name, i := range r.metrics {
		metrics[name] = i
	}
	return metrics
}
