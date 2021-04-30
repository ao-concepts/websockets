package websockets

// Logger interface
type Logger interface {
	Fatal(s string, args ...interface{})
	ErrError(err error)
	Error(s string, args ...interface{})
	ErrInfo(err error)
	Info(s string, args ...interface{})
	Debug(s string, args ...interface{})
}
