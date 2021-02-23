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
