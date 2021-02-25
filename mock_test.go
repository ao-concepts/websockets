package websockets_test

import (
	"sync"
	"testing"

	"github.com/ao-concepts/websockets"
	"github.com/stretchr/testify/assert"
)

func TestNewWebsocketConnMock(t *testing.T) {
	assert := assert.New(t)

	m := websockets.NewWebsocketConnMock()

	assert.NotNil(m)
	assert.NotNil(m.LocalData)
	assert.False(m.DoError)
}

func TestWebsocketConnMock_Locals(t *testing.T) {
	assert := assert.New(t)

	m := websockets.NewWebsocketConnMock()
	m.LocalData["test"] = "value"

	assert.Equal("value", m.Locals("test"))
}

func TestWebsocketConnMock_ReadJSON(t *testing.T) {
	assert := assert.New(t)
	wg := sync.WaitGroup{}

	m := websockets.NewWebsocketConnMock()

	// normal
	go func() {
		var data string

		assert.Nil(m.ReadJSON(&data))
		assert.Equal("test", data)
		wg.Done()
	}()

	wg.Add(1)
	m.Write <- "test"
	wg.Wait()

	// error
	m.DoError = true
	go func() {
		var data string

		assert.NotNil(m.ReadJSON(&data))
		wg.Done()
	}()

	wg.Add(1)
	m.Write <- "test2"
	wg.Wait()
}

func TestWebsocketConnMock_WriteJSON(t *testing.T) {
	assert := assert.New(t)
	wg := sync.WaitGroup{}

	m := websockets.NewWebsocketConnMock()

	// normal
	go func() {
		assert.Equal("test", <-m.Read)
		wg.Done()
	}()
	wg.Add(1)
	assert.Nil(m.WriteJSON("test"))
	wg.Wait()

	// error
	m.DoError = true
	assert.NotNil(m.WriteJSON("test"))
}

func TestWebsocketConnMock_Close(t *testing.T) {
	assert := assert.New(t)

	m := websockets.NewWebsocketConnMock()
	assert.Nil(m.Close())
	m.DoError = true
	assert.NotNil(m.Close())
}
