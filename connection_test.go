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
	c := websockets.NewConnection(s, nil)

	assert.NotPanics(func() {
		c.Publish(&websockets.Message{
			Event: "test",
			Payload: websockets.Payload{
				"value": "test-data",
			},
		}, nil)
	})
}

type wsConnection struct {
	locals map[string]interface{}
}

func (c *wsConnection) Locals(key string) interface{} {
	return c.locals[key]
}

func TestConnection_Locals(t *testing.T) {
	assert := assert.New(t)
	s, _ := websockets.New(nil, logging.New(logging.Debug, nil))
	conn := &wsConnection{
		locals: map[string]interface{}{
			"test-key": "test-value",
		},
	}
	c := websockets.NewConnection(s, conn)

	assert.Equal("test-value", c.Locals("test-key"))
	assert.Nil(c.Locals("not-available"))
}
