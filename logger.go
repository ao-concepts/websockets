package websockets

// Logger interface
type Logger interface {
	ErrError(err error)
	ErrInfo(err error)
	Info(s string, args ...interface{})
}
