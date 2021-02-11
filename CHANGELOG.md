# 0.2.0

*   (bug) Fix endless loop in connection write loop when a connection is closed.
*   (feature) Add `CountConnections` statistics function to `Server`.


# 0.1.1

*   (improvement) Run all subscription handler functions in a new go routine.


# 0.1.0

*   (bc) Updated signature of `Subscribe` metod on `Server` struct.


# 0.0.1

*   (feature) Abstraction layer for simple websocket handling.
