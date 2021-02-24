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
func (c *Connection) SendMessage(msg *Message) {
	c.write <- *msg
}

// Publish a message to all matching connections
func (c *Connection) Publish(msg *Message, filter Filter) {
	c.s.Publish(msg, filter)
}

// returns false on errors
func (c *Connection) listenForMessages(ctx context.Context, wc *websocket.Conn, onEnd func()) {
	defer func() {
		if r := recover(); r != nil {
			c.s.log.Warn("websocket: ending connection read after panic: %v", r)
		}

		onEnd()
	}()

	for {
		select {
		case <-ctx.Done():
			c.s.log.Info("Websocket closed: stopping reader")
			return
		default:
			var msg Message

			if err := wc.ReadJSON(&msg); err != nil {
				c.s.log.ErrInfo(err)
				return
			}

			msg.Connection = c

			if err := c.s.bus.Publish(msg.Event, &msg); err != nil {
				c.s.log.ErrError(err)
				return
			}
		}
	}

}

// publish messages to a websocket connection
func (c *Connection) publishMessages(ctx context.Context, wc *websocket.Conn, onEnd func()) {
	defer func() {
		if r := recover(); r != nil {
			c.s.log.Warn("websocket: ending connection write after panic: %v", r)
		}

		onEnd()
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
