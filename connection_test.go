package websockets_test

import (
	"net/url"
	"strconv"
	"testing"

	"github.com/ao-concepts/logging"
	"github.com/ao-concepts/websockets"
	"github.com/gofiber/fiber/v2"
	"github.com/gorilla/websocket"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
)

func startServer(assert *assert.Assertions) (ws *websockets.Server, port int) {
	app := fiber.New()
	log := logging.New(logging.Debug, nil)
	s, _ := websockets.New(nil, log)
	port, err := freeport.GetFreePort()
	assert.Nil(err)
	app.Get("/ws", s.Handler)
	go app.Listen(":" + strconv.Itoa(port))

	return s, port
}

func openConnection(port int, assert *assert.Assertions) *websocket.Conn {
	u := url.URL{Scheme: "ws", Host: "localhost:" + strconv.Itoa(port), Path: "/ws"}
	cc, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	assert.Nil(err)
	assert.NotNil(cc)

	return cc
}

func TestNewConnection(t *testing.T) {
	assert := assert.New(t)
	s, _ := websockets.New(nil, logging.New(logging.Debug, nil))
	c := websockets.NewConnection(s, nil)
	assert.NotNil(c)
}

func TestConnection_Set(t *testing.T) {
	assert := assert.New(t)
	s, _ := websockets.New(nil, logging.New(logging.Debug, nil))
	c := websockets.NewConnection(s, nil)

	c.Set("test-key", "test-value")
	assert.Equal("test-value", c.Get("test-key"))

	assert.Nil(c.Get("nil-key"))
}

func TestConnection_Get(t *testing.T) {
	assert := assert.New(t)
	s, _ := websockets.New(nil, logging.New(logging.Debug, nil))
	c := websockets.NewConnection(s, nil)

	assert.Nil(c.Get("test-key"))

	c.Set("test-key", "test-value")
	assert.Equal("test-value", c.Get("test-key"))
}

func TestConnection_Publish(t *testing.T) {
	assert := assert.New(t)
	s, _ := websockets.New(nil, logging.New(logging.Debug, nil))

	conn := websockets.NewConnection(s, nil)

	assert.NotPanics(func() {
		conn.Publish(&websockets.Message{
			Event: "test",
			Payload: websockets.Payload{
				"value": "test-data",
			},
		}, nil)
	})
}

func TestConnection_Locals(t *testing.T) {
	assert := assert.New(t)
	s, _ := websockets.New(nil, logging.New(logging.Debug, nil))
	wsConn := websockets.NewWebsocketConnMock()
	wsConn.LocalData["test-key"] = "test-value"

	conn := websockets.NewConnection(s, wsConn)

	assert.Equal("test-value", conn.Locals("test-key"))
	assert.Nil(conn.Locals("not-available"))
}
