package mock

import (
	"github.com/ao-concepts/websockets"
	"github.com/gofiber/fiber/v2"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

// Server mock websockets server
func Server(t *testing.T) (ws *websockets.Server, port int) {
	app := fiber.New()
	socket := websockets.New(NewServiceContainer())
	port, err := freeport.GetFreePort()
	assert.Nil(t, err)
	app.Get("/ws", socket.Handler)

	go func() {
		_ = app.Listen(":" + strconv.Itoa(port))
	}()

	return socket, port
}
