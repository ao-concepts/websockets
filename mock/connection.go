package mock

import (
	"context"
	"github.com/ao-concepts/websockets"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"net/url"
	"strconv"
	"testing"
)

// Connection mock
func Connection(port int, t *testing.T) *websocket.Conn {
	u := url.URL{Scheme: "ws", Host: "localhost:" + strconv.Itoa(port), Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	assert.Nil(t, err)
	assert.NotNil(t, conn)

	return conn
}

func WsConnection(wsConn *websockets.WebsocketConnMock) *websockets.Connection {
	ctx, cancel := context.WithCancel(context.Background())
	socket := websockets.New(NewServiceContainer())

	return websockets.NewConnection(socket, wsConn, ctx, cancel)
}
