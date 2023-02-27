package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAvailableSize(t *testing.T) {
	var total uint64 = 1000
	p := NewMemoryPool(total)

	assert.True(t, p.check())
	assert.Equal(t, total, p.GetAvailableSize())
	fetchID, err := p.PreFetch(10)
	assert.NoError(t, err)
	assert.Equal(t, fetchID, int64(101))
	assert.True(t, p.check())
	assert.Equal(t, total-uint64(10), p.GetAvailableSize())
}
