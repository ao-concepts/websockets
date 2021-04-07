package websockets

// Logger interface
type Logger interface {
	ErrError(err error)
	Error(s string, args ...interface{})
	Warn(s string, args ...interface{})
	ErrInfo(err error)
	Info(s string, args ...interface{})
	Debug(s string, args ...interface{})
}
