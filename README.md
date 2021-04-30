# ao-concepts websockets module

![CI](https://github.com/ao-concepts/websockets/workflows/CI/badge.svg)
[![codecov](https://codecov.io/gh/ao-concepts/websockets/branch/main/graph/badge.svg?token=AQVUZTRGQS)](https://codecov.io/gh/ao-concepts/websockets)

This module provides a websocket handler for usage with [fiber](https://github.com/gofiber/fiber). The handler is a abstraction layer that makes the handling of websocket messages easier.

## Contributing

If you are interested in contributing to this project, feel free to open an issue to discus a new feature, enhancement or improvement. If you found a bug or security vulnerability in this package, please start a issue, or open a PR against `master`.

## Installation

```shell
go get -u github.com/ao-concepts/websockets
```

## Usage

```go
cnt := NewServiceContainer() // the service container has to implement the `websockets.ServiceContainer` interface.
server, err := websockets.New(cnt)
if err != nil {
    log.ErrFatal(err)
}

// batch sent messages. This can be used to reduce load on clients.
// Batched message will be prefixed by `batch_`.
// The data will be stored as array of batched payloads below the `d` property of the actual sent message.
if err := server.UseBatch("event:name", time.Second); err != nil {
	log.ErrError(err)
}

server.Subscribe("event:name", func(msg *websockets.Message) {
    // here you can handle the message
    c := msg.Connection

    // access the payload
    fmt.Println(msg.Payload["value"])

    // you can set data to the current session
    c.Set("key", "value")
    c.Get("key")

    // you can respond directly on the connection that received the message.
    c.SendMessage(&websockets.Message{
        Event:   "event:name",
        Payload: websockets.Payload{
            "value": "data",
        },
    })

    // and you can respond to all connections that match a filter.
    c.SendMessage(&websockets.Message{
        Event:   "event:name",
        Payload: websockets.Payload{
            "value": "data",
        },
    }, func (c *websockets.Connection) bool {
        return c.Get("key") == true
    })
})

app := fiber.New()
app.Get("/ws", s.Handler)
app.Listen(":3000")
```

## Used packages 

This project uses some really great packages. Please make sure to check them out!

| Package                                                                   | Usage              |
| ------------------------------------------------------------------------- | ------------------ |
| [github.com/ao-concepts/eventbus](https://github.com/ao-concepts/storage) | Persistence helper |
| [github.com/go-co-op/gocron](https://github.com/go-co-op/gocron)          | Batch messages     |
| [github.com/gofiber/fiber](https://github.com/gofiber/fiber)                  | HTTP router        |
| [github.com/stretchr/testify](https://github.com/stretchr/testify)        | Testing            |
