# 0.11.5

*   (internal) Clean up context of connection.


# 0.11.4

*   (bug) Correctly close connections.


# 0.11.3

*   (bug) Fix adding of new connections.


# 0.11.2

*   (bug) Clean up locks.


# 0.11.1

*   (bug) Prevent deadlocks.


# 0.11.0

*   (bug) Fix issues with `UseBatch` cron task.
*   (bc) Pass ``time.Duration` to `UseBatch`.


# 0.10.1

*   (bug) Close connections in `Server.Shutdown` using go routines to prevent deadlocks.


# 0.10.0

*   (feature) Add `SendMessageUnBatched` to `Connection`.


# 0.9.0

*   (feature) Add `UseBatch` to batch messages of an event and publish them in a batched manner.
*   (bc) Improve performance by using smaller keys for the message object.


# 0.8.0

*   (feature) Add a filter to `CountConnections`.


# 0.7.8

*   (bug) Fix race condition when writing to client side ended connections (part 2).


# 0.7.7

*   (bug) Fix race condition when writing to client side ended connections.


# 0.7.6

*   (internal) Add debug logging.


# 0.7.5

*   (bug) Fix Always close context on ended connections.


# 0.7.1

*   (bug) Fix `Server.Connect` to accept `Connections`.


# 0.7.0

*   (feature) Make `Server.Connect` public for testing purposes.


# 0.6.0

*   (feature) Add `WebsocketConn` and a matching `WebsocketConnMock` for advanced test usage.


# 0.5.0

*   (bc) Introduce `Payload` type for `Message` payload.


# 0.4.7

*   (bug) Close connections correctly.


# 0.4.6

*   (bug) Recover and exit write loop of connection.


# 0.4.5

*   (bug) Do not try to close connections with errors.


# 0.4.4

*   (bug) Return on connection write errors.


# 0.4.3

*   (bug) Recover from panics in `onConnectionClose`.


# 0.4.2

*   (bug) Recover from panics in subscribers.


# 0.4.1

*   (bug) Do not add `Cnnection` to json of `Message`.


# 0.4.0

*   (feature) Add `Shutdown` to `Server`.
*   (feature) Add `SetOnConnectionClose` to `Server`.


# 0.3.0

*   (feature) Add `Locals` function to `Connection`.


# 0.2.0

*   (bug) Fix endless loop in connection write loop when a connection is closed.
*   (feature) Add `CountConnections` statistics function to `Server`.


# 0.1.1

*   (improvement) Run all subscription handler functions in a new go routine.


# 0.1.0

*   (bc) Updated signature of `Subscribe` metod on `Server` struct.


# 0.0.1

*   (feature) Abstraction layer for simple websocket handling.
