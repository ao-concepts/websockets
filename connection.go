package websockets

import (
	"context"
	"sync"

	"github.com/gofiber/websocket/v2"
)

// Connection on a websocket
type Connection struct {
	s     *Server
	wc    wsConnection
	data  map[string]interface{}
	write chan Message
	lock  sync.RWMutex
}

type wsConnection interface {
	Locals(key string) interface{}
}

// Filter function that checks if a connection matches some criteria
type Filter func(c *Connection) bool

// NewConnection constructor
func NewConnection(s *Server, wc wsConnection) *Connection {
	return &Connection{
		s:     s,
		wc:    wc,
		data:  make(map[string]interface{}),
		write: make(chan Message, s.config.WriteBufferSize),
		lock:  sync.RWMutex{},
	}
}

// Set a key on a connection to a specific value. Overwrites old value.
func (c *Connection) Set(key string, value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.data[key] = value
}

// Locals gets a local by key from the underlying connection.
func (c *Connection) Locals(key string) interface{} {
	return c.wc.Locals(key)
}

// Get the value of a key on the connection. Returns nil if the key does not exist
func (c *Connection) Get(key string) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if v, ok := c.data[key]; ok {
		return v
	}

	return nil
}

// SendMessage via this connection
func (c *Connection) SendMessage(m *Message) {
	c.write <- *m
}

// Publish a message to all matching connections
func (c *Connection) Publish(m *Message, filter Filter) {
	c.s.Publish(m, filter)
}

// returns false on errors
func (c *Connection) listenForMessages(wc *websocket.Conn) bool {
	var msg Message

	if err := wc.ReadJSON(&msg); err != nil {
		c.s.log.ErrInfo(err)

		return false
	}

	msg.Connection = c

	if err := c.s.bus.Publish(msg.Event, &msg); err != nil {
		c.s.log.ErrError(err)
		return false
	}

	return true
}

// publish messages to a websocket connection
func (c *Connection) publishMessages(ctx context.Context, wc *websocket.Conn) {
	defer func() {
		if r := recover(); r != nil {
			c.s.log.Warn("websocket: ending connection after panic: %v", r)
		}
	}()

	for {
		select {
		case msg := <-c.write:
			if err := wc.WriteJSON(msg); err != nil {
				c.s.log.ErrError(err)
				return
			}
		case <-ctx.Done():
			c.s.log.Info("Websocket closed: stopping writer")
			return
		}
	}
}
