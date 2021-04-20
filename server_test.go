// +build !race

package websockets_test

import (
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ao-concepts/logging"
	"github.com/ao-concepts/websockets"
	"github.com/dchest/uniuri"
	"github.com/gofiber/fiber/v2"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestNewServer(t *testing.T) {
	assert := assert.New(t)

	// missing logger
	s, err := websockets.New(nil, nil)
	assert.NotNil(err)

	// with logger
	s, err = websockets.New(nil, logging.New(logging.Debug, nil))
	assert.Nil(err)
	assert.NotNil(s)
}

func TestServer_Handler(t *testing.T) {
	assert := assert.New(t)
	app := fiber.New()
	s, _ := websockets.New(nil, logging.New(logging.Debug, nil))

	// no upgrade request
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	assert.NotNil(s.Handler(c))

	// upgrade request
	c = app.AcquireCtx(&fasthttp.RequestCtx{})
	c.Request().Header.Add("Connection", "upgrade")
	c.Request().Header.Add("Upgrade", "websocket")
	c.Request().Header.Add("Sec-Websocket-Version", "13")
	c.Request().Header.Add("Sec-WebSocket-Key", uniuri.NewLen(16))
	assert.Nil(s.Handler(c))

	// rest real dial
	s, port := startServer(assert)

	// no upgrade dial
	_, err := http.Get("localhost:" + strconv.Itoa(port) + "/ws")
	assert.NotNil(err)

	// successful dial
	openConnection(port, assert)
}

func TestServer_BrokenPipe(t *testing.T) {
	assert := assert.New(t)
	s, port := startServer(assert)
	ch := make(chan bool)
	wg := sync.WaitGroup{}

	s.SetOnConnectionClose(func(c *websockets.Connection) {
		wg.Done()
	})

	go func() {
		for {
			select {
			case <-ch:
				return
			default:
				s.Publish(&websockets.Message{
					Event: "test",
				}, nil)
			}
		}
	}()

	wg.Add(1)
	c := openConnection(port, assert)
	time.Sleep(10 * time.Millisecond)
	assert.Nil(c.Close())

	time.Sleep(10 * time.Millisecond)
	wg.Wait()

	wg.Add(1)
	c2 := openConnection(port, assert)
	assert.NotNil(c2)
	assert.Nil(c2.Close())
	ch <- true
	wg.Wait()
}

func TestServer_Subscribe(t *testing.T) {
	assert := assert.New(t)
	s, port := startServer(assert)
	time.Sleep(10 * time.Millisecond)
	c := openConnection(port, assert)
	wg := sync.WaitGroup{}

	assert.Nil(s.Subscribe("test", func(msg *websockets.Message) {
		assert.Equal("test-data", msg.Payload["value"])
		wg.Done()
	}))

	wg.Add(1)

	assert.Nil(c.WriteJSON(&websockets.Message{
		Event: "test",
		Payload: websockets.Payload{
			"value": "test-data",
		},
	}))

	wg.Wait()

	// test panic recovery
	assert.Nil(s.Subscribe("recover", func(msg *websockets.Message) {
		defer wg.Done()
		if msg.Payload["value"] == "panic" {
			panic("test")
		}
	}))

	wg.Add(1)
	assert.Nil(c.WriteJSON(&websockets.Message{
		Event: "recover",
		Payload: websockets.Payload{
			"value": "panic",
		},
	}))
	wg.Wait()

	wg.Add(1)
	assert.Nil(c.WriteJSON(&websockets.Message{
		Event: "recover",
		Payload: websockets.Payload{
			"value": "no-panic",
		},
	}))
	wg.Wait()

	assert.Nil(c.Close())
}

func TestServer_Publish(t *testing.T) {
	assert := assert.New(t)
	s, port := startServer(assert)
	c := openConnection(port, assert)
	wg := sync.WaitGroup{}

	wg.Add(1)

	go func(c *websocket.Conn) {
		var msg websockets.Message
		assert.Nil(c.ReadJSON(&msg))
		assert.Equal("test-data", msg.Payload["value"])
		wg.Done()
	}(c)

	s.Publish(&websockets.Message{
		Event: "test",
		Payload: websockets.Payload{
			"value": "test-data",
		},
	}, nil)

	wg.Wait()

	// filter (false)
	c2 := openConnection(port, assert)

	go func(c *websocket.Conn) {
		var msg websockets.Message
		assert.Nil(c.ReadJSON(&msg))
		assert.NotNil(c.ReadJSON(&msg))
	}(c)

	go func(c *websocket.Conn) {
		var msg websockets.Message
		assert.Nil(c.ReadJSON(&msg))
		assert.NotNil(c.ReadJSON(&msg))
	}(c2)

	s.Publish(&websockets.Message{
		Event: "filter-test",
		Payload: websockets.Payload{
			"value": "filter-test-data",
		},
	}, func(c *websockets.Connection) bool {
		return false
	})

	// filter (false)
	s.Publish(&websockets.Message{
		Event: "filter-test",
		Payload: websockets.Payload{
			"value": "filter-test-data",
		},
	}, func(c *websockets.Connection) bool {
		return true
	})

	time.Sleep(10 * time.Millisecond)

	assert.Nil(c.Close())
	assert.Nil(c2.Close())
}

func TestServer_CountConnections(t *testing.T) {
	assert := assert.New(t)

	s, port := startServer(assert)

	assert.Equal(0, s.CountConnections(nil))

	c := openConnection(port, assert)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(1, s.CountConnections(nil))

	c2 := openConnection(port, assert)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(2, s.CountConnections(nil))
	assert.Equal(0, s.CountConnections(func(c *websockets.Connection) bool {
		return false
	}))
	assert.Equal(2, s.CountConnections(func(c *websockets.Connection) bool {
		return true
	}))

	assert.Nil(c.Close())
	time.Sleep(100 * time.Millisecond)
	assert.Equal(1, s.CountConnections(nil))

	assert.Nil(c2.Close())
	time.Sleep(100 * time.Millisecond)
	assert.Equal(0, s.CountConnections(nil))
}

func TestServer_SetOnConnectionClose(t *testing.T) {
	assert := assert.New(t)

	s, port := startServer(assert)
	c := openConnection(port, assert)

	wg := sync.WaitGroup{}

	s.SetOnConnectionClose(func(c *websockets.Connection) {
		wg.Done()
	})

	wg.Add(1)
	assert.Nil(c.Close())
	wg.Wait()

	// test panic
	c2 := openConnection(port, assert)

	s.SetOnConnectionClose(func(c *websockets.Connection) {
		defer wg.Done()
		panic("test")
	})

	wg.Add(1)
	assert.Nil(c2.Close())
	wg.Wait()
}

func TestServer_Shutdown(t *testing.T) {
	assert := assert.New(t)

	s, port := startServer(assert)
	openConnection(port, assert)
	openConnection(port, assert)
	openConnection(port, assert)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(3, s.CountConnections(nil))

	wg := sync.WaitGroup{}

	s.SetOnConnectionClose(func(c *websockets.Connection) {
		wg.Done()
	})

	assert.Nil(s.Subscribe("test", func(msg *websockets.Message) {}))
	assert.Nil(s.Subscribe("test-2", func(msg *websockets.Message) {}))

	wg.Add(3)
	s.Shutdown()
	wg.Wait()

	assert.Equal(0, s.CountConnections(nil))
}

func TestServer_UseBatch(t *testing.T) {
	assert := assert.New(t)
	s, port := startServer(assert)
	assert.Nil(s.UseBatch("test", time.Second))

	c := openConnection(port, assert)
	wg := sync.WaitGroup{}

	go func(c *websocket.Conn) {
		var msg websockets.Message
		assert.Nil(c.ReadJSON(&msg))
		assert.Equal("batch-test", msg.Event)
		assert.Len(msg.Payload["d"], 2)
		wg.Done()
	}(c)

	wg.Add(1)
	s.Publish(&websockets.Message{
		Event: "test",
		Payload: websockets.Payload{
			"value": "test-data",
		},
	}, nil)
	s.Publish(&websockets.Message{
		Event: "test",
		Payload: websockets.Payload{
			"value": "test-data",
		},
	}, nil)

	wg.Wait()
	assert.Nil(c.Close())
}
