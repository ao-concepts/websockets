package websockets_test

import (
	"context"
	"github.com/ao-concepts/websockets"
	"github.com/ao-concepts/websockets/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewConnection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	socket := websockets.New(mock.NewServiceContainer())

	assert.NotNil(t, websockets.NewConnection(socket, nil, ctx, cancel))
}

func TestConnection_Set(t *testing.T) {
	conn := mock.WsConnection(nil)
	conn.Set("test-key", "test-value")

	assert.Equal(t, "test-value", conn.Get("test-key"))
	assert.Nil(t, conn.Get("nil-key"))
}

func TestConnection_Get(t *testing.T) {
	conn := mock.WsConnection(nil)
	assert.Nil(t, conn.Get("test-key"))

	conn.Set("test-key", "test-value")
	assert.Equal(t, "test-value", conn.Get("test-key"))
}

func TestConnection_Publish(t *testing.T) {
	conn := mock.WsConnection(nil)

	assert.NotPanics(t, func() {
		conn.Publish(&websockets.Message{
			Event: "test",
			Payload: websockets.Payload{
				"value": "test-data",
			},
		}, nil)
	})
}

func TestConnection_Locals(t *testing.T) {
	wsConn := websockets.NewWebsocketConnMock()
	wsConn.LocalData["test-key"] = "test-value"

	conn := mock.WsConnection(wsConn)

	assert.Equal(t, "test-value", conn.Locals("test-key"))
	assert.Nil(t, conn.Locals("not-available"))
}
