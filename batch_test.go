package websockets_test

import (
	"github.com/ao-concepts/websockets"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewBatch(t *testing.T) {
	b := websockets.NewBatch()
	assert.NotNil(t, b)
}

func TestBatch_AddPayload_GetDataAndRemoveAll(t *testing.T) {
	b := websockets.NewBatch()

	assert.Len(t, b.GetDataAndRemoveAll(), 0)

	p1 := websockets.Payload{
		"value": "test",
	}
	b.AddPayload(p1)
	p2 := websockets.Payload{
		"value": "test2",
	}
	b.AddPayload(p2)

	data := b.GetDataAndRemoveAll()
	assert.Len(t, data, 2)
	assert.Equal(t, p1, data[0])
	assert.Equal(t, p2, data[1])
	assert.Len(t, b.GetDataAndRemoveAll(), 0)
}
