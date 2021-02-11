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

func TestServer_Subscribe(t *testing.T) {
	assert := assert.New(t)
	s, port := startServer(assert)
	c := openConnection(port, assert)
	wg := sync.WaitGroup{}

	assert.Nil(s.Subscribe("test", func(msg *websockets.Message) {
		assert.Equal("test-data", msg.Payload)
		wg.Done()
	}))

	wg.Add(1)

	assert.Nil(c.WriteJSON(&websockets.Message{
		Event:   "test",
		Payload: "test-data",
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
		assert.Equal("test-data", msg.Payload)
		wg.Done()
	}(c)

	s.Publish(&websockets.Message{
		Event:   "test",
		Payload: "test-data",
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
		Event:   "filter-test",
		Payload: "filter-test-data",
	}, func(c *websockets.Connection) bool {
		return false
	})

	// filter (false)
	s.Publish(&websockets.Message{
		Event:   "filter-test",
		Payload: "filter-test-data",
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

	assert.Equal(0, s.CountConnections())

	c := openConnection(port, assert)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(1, s.CountConnections())

	c2 := openConnection(port, assert)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(2, s.CountConnections())

	assert.Nil(c.Close())
	time.Sleep(100 * time.Millisecond)
	assert.Equal(1, s.CountConnections())

	assert.Nil(c2.Close())
	time.Sleep(100 * time.Millisecond)
	assert.Equal(0, s.CountConnections())
}
