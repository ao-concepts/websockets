# ao-concepts websockets module

![CI](https://github.com/ao-concepts/websockets/workflows/CI/badge.svg)
[![codecov](https://codecov.io/gh/ao-concepts/websockets/branch/master/graph/badge.svg?token=AQVUZTRGQS)](https://codecov.io/gh/ao-concepts/websockets)

This module provides a websocket handler for usage with [fiber](https://github.com/gofiber/fiber). The handler is a abstraction layer that makes the handling of websocket messages easier.

## Information

The ao-concepts ecosystem is still under active development and therefore the API of this module may have breaking changes until there is a first stable release.

If you are interested in contributing to this project, feel free to open a issue to discus a new feature, enhancement or improvement. If you found a bug or security vulnerability in this package, please start a issue, or open a PR against `master`.

## Installation

```shell
go get -u github.com/ao-concepts/websockets
```

## Usage

```go
log := logging.New(logging.Debug, nil)
server, err := websockets.New(nil, log)
if err != nil {
    log.ErrFatal(err)
}

ch := make(chan eventbus.Event) // @todo: this should be some kind of internal
if err := server.Subscribe("event:name", ch); err != nil {
    log.ErrFatal(err)
}

go func () {
    for  {
        msg := <- ch
        fmt.Println(msg)
        // here you can handle the message
        c := msg.Connection

        // you can set data to the current session
        c.Set("key", "value")
        c.Get("key")

        // you can respond directly on the connection that received the message.
        c.SendMessage(&websockets.Message{
            Event:   "event:name",
            Payload: "data",
        })

        // and you can respond to all connections that match a filter.
        c.SendMessage(&websockets.Message{
            Event:   "event:name",
            Payload: "data",
        }, func (c *websockets.Connection) bool {
            return c.Get("key") == true
        })
    }
}()

app := fiber.New()
app.Get("/ws", s.Handler)
app.Listen(":3000")
```

## Used packages 

This project uses some really great packages. Please make sure to check them out!

| Package                                                                   | Usage              |
| ------------------------------------------------------------------------- | ------------------ |
| [github.com/ao-concepts/eventbus](https://github.com/ao-concepts/storage) | Persistence helper |
| [github.com/gofiber/fiber](https://github.com/gofiber/fiber)              | HTTP router        |
| [github.com/stretchr/testify](https://github.com/stretchr/testify)        | Testing            |
