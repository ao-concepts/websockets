package websockets

import (
	"context"
	"fmt"
	"github.com/ao-concepts/eventbus"
	"github.com/go-co-op/gocron"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"sync"
	"time"
)

// Message that is received or sent via a websocket.
type Message struct {
	Event      string      `json:"e"`
	Payload    Payload     `json:"p"`
	Connection *Connection `json:"-"`
}

// Payload send by a websocket connection
type Payload map[string]interface{}

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
	batches           map[string]bool
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
		config:  config,
		log:     log,
		bus:     eventbus.New(nil, false),
		lock:    sync.RWMutex{},
		ctx:     ctx,
		cancel:  cancel,
		batches: make(map[string]bool),
	}, nil
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown() {
	s.lock.Lock()
	wg := sync.WaitGroup{}

	for _, conn := range s.connections {
		if s.onConnectionClose != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				s.onConnectionClose(conn)
			}()
		}
	}
	s.lock.Unlock()

	wg.Wait()

	s.lock.Lock()
	defer s.lock.Unlock()

	s.cancel()
	s.connections = nil
	s.isStopped = true
}

// Handler that reads from and writes to a websocket connection
func (s *Server) Handler(c *fiber.Ctx) error {
	return websocket.New(func(wc *websocket.Conn) {
		conn := NewConnection(s, wc)
		s.addConnection(conn)
		s.Connect(conn)
	})(c)
}

// Subscribe to a websocket event
func (s *Server) Subscribe(eventName string, handler func(msg *Message)) error {
	ch := make(chan eventbus.Event)

	go s.handleSubscription(ch, handler)

	return s.bus.Subscribe(eventName, ch)
}

// Publish data to all matching connections
func (s *Server) Publish(msg *Message, filter Filter) {
	var connections []*Connection

	s.lock.RLock()
	connections = append(connections, s.connections...)
	s.lock.RUnlock()

	go publishToConnections(msg, filter, connections)
}

// UseBatch registers a batch interval sender.
// Sends events in a batched manner when they are sent by `Connection.SendMessage`.
// Has to be called before the first connection is established.
// Interval in seconds.
func (s *Server) UseBatch(event string, interval time.Duration) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	eventName := event
	s.batches[eventName] = true

	sc := gocron.NewScheduler(time.UTC)

	if _, err := sc.Every(interval).SingletonMode().Do(func() {
		s.lock.RLock()
		connections := s.connections
		s.lock.RUnlock()

		wg := sync.WaitGroup{}

		for _, c := range connections {
			wg.Add(1)

			go func(c *Connection) {
				defer wg.Done()
				c.sendBatch(eventName)
			}(c)
		}

		wg.Wait()
	}); err != nil {
		return err
	}

	sc.StartAsync()
	return nil
}

// GetBatches return the current list of batches registered for the server.
func (s *Server) GetBatches() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var batches []string

	for event := range s.batches {
		batches = append(batches, event)
	}

	return batches
}

func publishToConnections(msg *Message, filter Filter, connections []*Connection) {
	if filter == nil {
		for _, c := range connections {
			c.SendMessage(msg)
		}

		return
	}

	for _, c := range connections {
		if filter(c) {
			c.SendMessage(msg)
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

// CountConnections returns the number of currently active connections (a filter can be used to restrict the count)
func (s *Server) CountConnections(filter Filter) int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if filter == nil {
		return len(s.connections)
	}

	counter := 0

	for _, c := range s.connections {
		if filter(c) {
			counter++
		}
	}

	return counter
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

// Connect a websocket to the server
func (s *Server) Connect(conn *Connection) {
	if s.isStopped {
		conn.wc.Close()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	endCalled := false
	sync := sync.Mutex{}
	onEnd := func() {
		sync.Lock()
		defer sync.Unlock()
		defer cancel()

		if endCalled {
			return
		}

		endCalled = true

		if s.onConnectionClose != nil {
			s.onConnectionClose(conn)
		}

		s.removeConnection(conn)
	}

	go conn.publishMessages(ctx, conn.wc, onEnd)
	conn.listenForMessages(ctx, conn.wc, onEnd)
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
