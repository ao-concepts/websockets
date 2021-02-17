package websockets

import (
	"context"
	"fmt"
	"sync"

	"github.com/ao-concepts/eventbus"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// Message that is received or sent via a websocket.
type Message struct {
	Event      string      `json:"event"`
	Payload    interface{} `json:"payload"`
	Connection *Connection `json:"-"`
}

// Server for websockets
type Server struct {
	lock              sync.RWMutex
	log               Logger
	bus               *eventbus.Bus
	connections       []*Connection
	config            *ServerConfig
	onConnectionClose OnConnectionClose
	ctx               context.Context
	cancel            context.CancelFunc
	isStopped         bool
}

// ServerConfig configuration of the server
type ServerConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
}

// OnConnectionClose is ecuted when a connection is closed
type OnConnectionClose func(c *Connection)

// New server constructor
func New(config *ServerConfig, log Logger) (s *Server, err error) {
	if log == nil {
		return nil, fmt.Errorf("Cannot use a websocket server without a logger")
	}

	if config == nil {
		config = &ServerConfig{
			ReadBufferSize:  5,
			WriteBufferSize: 5,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		config: config,
		log:    log,
		bus:    eventbus.New(nil, false),
		lock:   sync.RWMutex{},
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Shutdown gracefully stopps the server
func (s *Server) Shutdown() {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, conn := range s.connections {
		if s.onConnectionClose != nil {
			s.onConnectionClose(conn)
		}
	}

	s.cancel()
	s.connections = nil
	s.isStopped = true
}

// Handler that reads from and writes to a websocket connection
func (s *Server) Handler(c *fiber.Ctx) error {
	return websocket.New(func(wc *websocket.Conn) {
		s.connect(wc)
	})(c)
}

// Subscribe to a websocket event
func (s *Server) Subscribe(eventName string, handler func(msg *Message)) error {
	ch := make(chan eventbus.Event)

	go s.handleSubscription(ch, handler)

	return s.bus.Subscribe(eventName, ch)
}

// Publish data to all matching connections
func (s *Server) Publish(m *Message, filter Filter) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if filter == nil {
		for _, c := range s.connections {
			c.SendMessage(m)
		}

		return
	}

	for _, c := range s.connections {
		if filter(c) {
			c.SendMessage(m)
		}
	}
}

// SetOnConnectionClose sets the function that is executed when a connection is closed
func (s *Server) SetOnConnectionClose(fn OnConnectionClose) {
	s.onConnectionClose = func(c *Connection) {
		defer func() {
			if r := recover(); r != nil {
				s.log.Error("websocket: recovering connection close from panic: %v", r)
			}
		}()

		fn(c)
	}
}

// CountConnections returns the number of currently active connections
func (s *Server) CountConnections() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.connections)
}

func (s *Server) addConnection(conn *Connection) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.connections = append(s.connections, conn)
}

func (s *Server) removeConnection(conn *Connection) {
	s.lock.Lock()
	defer s.lock.Unlock()

	ci := -1

	for i, c := range s.connections {
		if c == conn {
			ci = i
		}
	}

	if ci >= 0 {
		l := len(s.connections) - 1
		s.connections[l], s.connections[ci] = s.connections[ci], s.connections[l]
		s.connections = s.connections[:l]
	}
}

// connect a websocket to the server
func (s *Server) connect(wc *websocket.Conn) {
	if s.isStopped {
		wc.Close()
		return
	}

	conn := NewConnection(s, wc)
	s.addConnection(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go conn.publishMessages(ctx, wc)

	for {
		if !conn.listenForMessages(wc) {
			if s.onConnectionClose != nil {
				s.onConnectionClose(conn)
			}

			s.removeConnection(conn)
			break
		}
	}
}

func (s *Server) handleSubscription(ch chan eventbus.Event, handler func(msg *Message)) {
	for {
		select {
		case msg := <-ch:
			go func() {
				defer func() {
					if r := recover(); r != nil {
						s.log.Error("websocket: recovering subscription from panic: %v", r)
					}
				}()

				handler(msg.Data.(*Message))
			}()
		case <-s.ctx.Done():
			return
		}
	}
}
