// +build !race

package websockets_test

import (
	"github.com/ao-concepts/websockets"
	"github.com/ao-concepts/websockets/mock"
	"github.com/dchest/uniuri"
	"github.com/gofiber/fiber/v2"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	// missing logger
	assert.Panics(t, func() {
		websockets.New(&mock.ServiceContainer{})
	})

	// with logger
	assert.NotNil(t, websockets.New(mock.NewServiceContainer()))
}

func TestServer_Handler(t *testing.T) {
	app := fiber.New()
	socket := websockets.New(mock.NewServiceContainer())

	// no upgrade request
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	assert.NotNil(t, socket.Handler(ctx))

	// upgrade request
	ctx = app.AcquireCtx(&fasthttp.RequestCtx{})
	ctx.Request().Header.Add("Connection", "upgrade")
	ctx.Request().Header.Add("Upgrade", "websocket")
	ctx.Request().Header.Add("Sec-Websocket-Version", "13")
	ctx.Request().Header.Add("Sec-WebSocket-Key", uniuri.NewLen(16))
	assert.Nil(t, socket.Handler(ctx))

	// rest real dial
	socket, port := mock.Server(t)

	// no upgrade dial
	_, err := http.Get("localhost:" + strconv.Itoa(port) + "/ws")
	assert.NotNil(t, err)

	// successful dial
	mock.Connection(port, t)
}

func TestServer_BrokenPipe(t *testing.T) {
	socket, port := mock.Server(t)
	ch := make(chan bool)
	wg := sync.WaitGroup{}

	socket.SetOnConnectionClose(func(c *websockets.Connection) {
		wg.Done()
	})

	go func() {
		for {
			select {
			case <-ch:
				return
			default:
				socket.Publish(&websockets.Message{
					Event: "test",
				}, nil)
			}
		}
	}()

	wg.Add(1)
	conn := mock.Connection(port, t)
	time.Sleep(10 * time.Millisecond)
	assert.Nil(t, conn.Close())

	time.Sleep(10 * time.Millisecond)
	wg.Wait()

	wg.Add(1)
	conn2 := mock.Connection(port, t)
	assert.NotNil(t, conn2)
	assert.Nil(t, conn2.Close())
	ch <- true
	wg.Wait()
}

func TestServer_Subscribe(t *testing.T) {
	socket, port := mock.Server(t)
	time.Sleep(10 * time.Millisecond)
	conn := mock.Connection(port, t)
	wg := sync.WaitGroup{}

	assert.Nil(t, socket.Subscribe("test", func(msg *websockets.MessageWithConnection) {
		assert.Equal(t, "test-data", msg.Payload["value"])
		wg.Done()
	}))

	wg.Add(1)

	assert.Nil(t, conn.WriteJSON(&websockets.Message{
		Event: "test",
		Payload: websockets.Payload{
			"value": "test-data",
		},
	}))

	wg.Wait()

	// test panic recovery
	assert.Nil(t, socket.Subscribe("recover", func(msg *websockets.MessageWithConnection) {
		defer wg.Done()
		if msg.Payload["value"] == "panic" {
			panic("test")
		}
	}))

	wg.Add(1)
	assert.Nil(t, conn.WriteJSON(&websockets.Message{
		Event: "recover",
		Payload: websockets.Payload{
			"value": "panic",
		},
	}))
	wg.Wait()

	wg.Add(1)
	assert.Nil(t, conn.WriteJSON(&websockets.Message{
		Event: "recover",
		Payload: websockets.Payload{
			"value": "no-panic",
		},
	}))
	wg.Wait()

	assert.Nil(t, conn.Close())
}

func TestServer_Publish(t *testing.T) {
	socket, port := mock.Server(t)
	conn := mock.Connection(port, t)
	wg := sync.WaitGroup{}

	wg.Add(1)

	go func(c *websocket.Conn) {
		var msg websockets.Message
		assert.Nil(t, c.ReadJSON(&msg))
		assert.Equal(t, "test-data", msg.Payload["value"])
		wg.Done()
	}(conn)

	socket.Publish(&websockets.Message{
		Event: "test",
		Payload: websockets.Payload{
			"value": "test-data",
		},
	}, nil)

	wg.Wait()

	// filter (false)
	conn2 := mock.Connection(port, t)

	go func(conn *websocket.Conn) {
		var msg websockets.Message
		assert.Nil(t, conn.ReadJSON(&msg))
		assert.NotNil(t, conn.ReadJSON(&msg))
	}(conn)

	go func(conn *websocket.Conn) {
		var msg websockets.Message
		assert.Nil(t, conn.ReadJSON(&msg))
		assert.NotNil(t, conn.ReadJSON(&msg))
	}(conn2)

	socket.Publish(&websockets.Message{
		Event: "filter-test",
		Payload: websockets.Payload{
			"value": "filter-test-data",
		},
	}, func(c *websockets.Connection) bool {
		return false
	})

	// filter (false)
	socket.Publish(&websockets.Message{
		Event: "filter-test",
		Payload: websockets.Payload{
			"value": "filter-test-data",
		},
	}, func(c *websockets.Connection) bool {
		return true
	})

	time.Sleep(10 * time.Millisecond)

	assert.Nil(t, conn.Close())
	assert.Nil(t, conn2.Close())
}

func TestServer_CountConnections(t *testing.T) {
	socket, port := mock.Server(t)
	assert.Equal(t, 0, socket.CountConnections(nil))

	conn := mock.Connection(port, t)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, socket.CountConnections(nil))

	conn2 := mock.Connection(port, t)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 2, socket.CountConnections(nil))
	assert.Equal(t, 0, socket.CountConnections(func(conn *websockets.Connection) bool {
		return false
	}))
	assert.Equal(t, 2, socket.CountConnections(func(conn *websockets.Connection) bool {
		return true
	}))

	assert.Nil(t, conn.Close())
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, socket.CountConnections(nil))

	assert.Nil(t, conn2.Close())
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 0, socket.CountConnections(nil))
}

func TestServer_SetOnConnectionClose(t *testing.T) {
	socket, port := mock.Server(t)
	conn := mock.Connection(port, t)

	wg := sync.WaitGroup{}

	socket.SetOnConnectionClose(func(c *websockets.Connection) {
		wg.Done()
	})

	wg.Add(1)
	assert.Nil(t, conn.Close())
	wg.Wait()

	// test panic
	conn2 := mock.Connection(port, t)

	socket.SetOnConnectionClose(func(c *websockets.Connection) {
		defer wg.Done()
		panic("test")
	})

	wg.Add(1)
	assert.Nil(t, conn2.Close())
	wg.Wait()
}

func TestServer_Shutdown(t *testing.T) {
	socket, port := mock.Server(t)
	mock.Connection(port, t)
	mock.Connection(port, t)
	mock.Connection(port, t)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 3, socket.CountConnections(nil))

	wg := sync.WaitGroup{}

	socket.SetOnConnectionClose(func(c *websockets.Connection) {
		wg.Done()
	})

	assert.Nil(t, socket.Subscribe("test", func(msg *websockets.MessageWithConnection) {}))
	assert.Nil(t, socket.Subscribe("test-2", func(msg *websockets.MessageWithConnection) {}))

	wg.Add(3)
	socket.Shutdown()
	wg.Wait()

	assert.Equal(t, 0, socket.CountConnections(nil))

	socket.Connect(websockets.NewConnection(socket, nil, nil, func() {}))

	assert.Equal(t, 0, socket.CountConnections(nil))
}

func TestServer_UseBatch(t *testing.T) {
	socket, port := mock.Server(t)
	assert.Nil(t, socket.UseBatch("test", time.Second))

	conn := mock.Connection(port, t)
	wg := sync.WaitGroup{}

	go func(conn *websocket.Conn) {
		var msg websockets.Message
		assert.Nil(t, conn.ReadJSON(&msg))
		assert.Equal(t, "batch-test", msg.Event)
		assert.Len(t, msg.Payload["d"], 2)
		wg.Done()
	}(conn)

	wg.Add(1)
	socket.Publish(&websockets.Message{
		Event: "test",
		Payload: websockets.Payload{
			"value": "test-data",
		},
	}, nil)
	socket.Publish(&websockets.Message{
		Event: "test",
		Payload: websockets.Payload{
			"value": "test-data",
		},
	}, nil)

	wg.Wait()
	assert.Nil(t, conn.Close())
}
