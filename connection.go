package websockets

import (
	"context"
	"github.com/gofiber/websocket/v2"
	"sync"
	"time"
)

// Connection on a websocket
type Connection struct {
	lock      sync.RWMutex
	s         *Server
	wc        WebsocketConn
	data      map[string]interface{}
	write     chan Message
	batches   map[string]*Batch
	isOpen    bool
	ctx       context.Context
	cancelCtx func()
}

// WebsocketConn websocket connection interface
type WebsocketConn interface {
	ReadJSON(v interface{}) error
	WriteJSON(v interface{}) error
	WriteMessage(messageType int, data []byte) error
	Close() error
	Locals(key string) interface{}
}

// Filter function that checks if a connection matches some criteria
type Filter func(c *Connection) bool

// OnEndFunc function that is executed on connection ends
type OnEndFunc = func(c *Connection)

// NewConnection constructor
func NewConnection(s *Server, wc WebsocketConn, ctx context.Context, cancel func()) *Connection {
	c := &Connection{
		s:         s,
		wc:        wc,
		data:      make(map[string]interface{}),
		write:     make(chan Message, s.cfg.WriteBufferSize),
		lock:      sync.RWMutex{},
		batches:   make(map[string]*Batch),
		isOpen:    true,
		ctx:       ctx,
		cancelCtx: cancel,
	}

	for _, event := range s.GetBatches() {
		c.batches[event] = NewBatch()
	}

	return c
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

// SendMessage via this connection. Message can be sent batched using `UseBatch`.
func (c *Connection) SendMessage(msg *Message) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if _, ok := c.batches[msg.Event]; ok {
		c.batches[msg.Event].AddPayload(msg.Payload)
		return
	}

	if !c.isOpen {
		return
	}

	go func() {
		c.write <- *msg
	}()
}

// SendMessageUnBatched send a message via this connection. Massage will not be batched.
func (c *Connection) SendMessageUnBatched(msg *Message) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if !c.isOpen {
		return
	}

	go func() {
		c.write <- *msg
	}()
}

// Publish a message to all matching connections
func (c *Connection) Publish(msg *Message, filter Filter) {
	c.s.Publish(msg, filter)
}

// returns true if the connection was open before
func (c *Connection) close() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	success := c.isOpen
	c.isOpen = false

	return success
}

// send all batched data for a event
func (c *Connection) sendBatch(event string) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if batch, ok := c.batches[event]; ok {
		data := batch.GetDataAndRemoveAll()

		if len(data) == 0 {
			return
		}

		go c.SendMessageUnBatched(&Message{
			Event: "batch-" + event,
			Payload: Payload{
				"d": data,
			},
		})
	}
}

// returns false on errors
func (c *Connection) listenForMessages(wc WebsocketConn) {
	defer func() {
		if r := recover(); r != nil {
			c.s.log.Error("websockets: ending connection read after panic: %v", r)
		}
	}()
	defer c.cancelCtx()

	for {
		select {
		case <-c.ctx.Done():
			c.s.log.Info("websockets: websocket closed - stopping reader")
			return
		default:
			var msg MessageWithConnection

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
func (c *Connection) publishMessages(wc WebsocketConn) {
	ticker := time.NewTicker(PingInterval)
	defer func() {
		if r := recover(); r != nil {
			c.s.log.Error("websockets: ending connection write after panic: %v", r)
		}
	}()
	defer c.cancelCtx()

	for {
		select {
		case msg := <-c.write:
			if err := wc.WriteJSON(msg); err != nil {
				c.s.log.ErrError(err)
				return
			}
		case <-ticker.C:
			if err := wc.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case <-c.ctx.Done():
			c.s.log.Info("websockets: websocket closed - stopping writer")
			return
		}
	}
}
