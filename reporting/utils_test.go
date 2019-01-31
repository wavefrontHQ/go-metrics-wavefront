package reporting

import (
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
