package websockets_test

import (
	"fmt"
	"net/url"
	"strconv"
	"testing"

	"github.com/ao-concepts/websockets"
	"github.com/gofiber/fiber/v2"
	"github.com/gorilla/websocket"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
)

type testLogger struct {
	logs []string
}

func (l *testLogger) ErrError(err error) {
	fmt.Println(err)
	l.logs = append(l.logs, err.Error())
}

func (l *testLogger) ErrInfo(err error) {
	fmt.Println(err)
	l.logs = append(l.logs, err.Error())
}

func (l *testLogger) Info(s string, args ...interface{}) {
	err := fmt.Sprintf(s, args...)
	fmt.Println(err)
	l.logs = append(l.logs, err)
}

func startServer(assert *assert.Assertions) (ws *websockets.Server, port int) {
	app := fiber.New()
	s, _ := websockets.New(nil, &testLogger{})
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
	s, _ := websockets.New(nil, &testLogger{})
	c := websockets.NewConnection(s)
	assert.NotNil(c)
}

func TestConnection_Set(t *testing.T) {
	assert := assert.New(t)
	s, _ := websockets.New(nil, &testLogger{})
	c := websockets.NewConnection(s)

	c.Set("test-key", "test-value")
	assert.Equal("test-value", c.Get("test-key"))

	assert.Nil(c.Get("nil-key"))
}

func TestConnection_Get(t *testing.T) {
	assert := assert.New(t)
	s, _ := websockets.New(nil, &testLogger{})
	c := websockets.NewConnection(s)

	assert.Nil(c.Get("test-key"))

	c.Set("test-key", "test-value")
	assert.Equal("test-value", c.Get("test-key"))
}

func TestConnection_Publish(t *testing.T) {
	assert := assert.New(t)
	s, _ := websockets.New(nil, &testLogger{})
	c := websockets.NewConnection(s)

	assert.NotPanics(func() {
		c.Publish(&websockets.Message{
			Event:   "test",
			Payload: "test-data",
		}, nil)
	})
}
