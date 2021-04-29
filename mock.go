package websockets

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

// WebsocketConnMock mock struct for websocket connections
// You can set `DoError` to let any method that can return an error return errors
type WebsocketConnMock struct {
	Read      chan interface{}
	Write     chan interface{}
	LocalData map[string]interface{}
	DoError   bool
}

// NewWebsocketConnMock constructor
func NewWebsocketConnMock() *WebsocketConnMock {
	return &WebsocketConnMock{
		Read:      make(chan interface{}),
		Write:     make(chan interface{}),
		LocalData: make(map[string]interface{}),
		DoError:   false,
	}
}

// Locals access local data
func (c *WebsocketConnMock) Locals(key string) interface{} {
	return c.LocalData[key]
}

// ReadJSON receives data from the Write chan
func (c *WebsocketConnMock) ReadJSON(v interface{}) error {
	data := <-c.Write

	if c.DoError {
		return fmt.Errorf("error")
	}

	return mapstructure.Decode(data, v)
}

// WriteJSON writes data to the Read chan
func (c *WebsocketConnMock) WriteJSON(v interface{}) error {
	if c.DoError {
		return fmt.Errorf("error")
	}

	c.Read <- v

	return nil
}

// Close the connection
func (c *WebsocketConnMock) Close() error {
	if c.DoError {
		return fmt.Errorf("error")
	}

	return nil
}

// WriteMessage to the connection
func (c *WebsocketConnMock) WriteMessage(messageType int, data []byte) error {
	if c.DoError {
		return fmt.Errorf("error")
	}

	c.Read <- data

	return nil
}
