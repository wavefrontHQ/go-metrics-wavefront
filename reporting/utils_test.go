package reporting

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMetricKey test EncodeKey and DecodeKey
func TestMetricKey(t *testing.T) {
	name := "met.ri[]c na$&me"
	tags := map[string]string{"env": "test", "k=e&y[pp]": "val=ue&va[lu]e"}

	key := EncodeKey(name, tags)

	newName, newTags := DecodeKey(key)

	assert.Equal(t, newName, name)
	assert.Equal(t, newTags, tags)
}

func TestEmptyTags(t *testing.T) {
	key := EncodeKey("metric.name", nil)
	assert.False(t, strings.Contains(key, "["))

	s, tags := DecodeKey(key)
	assert.True(t, len(tags) == 0)
	assert.True(t, s == "metric.name")
}
