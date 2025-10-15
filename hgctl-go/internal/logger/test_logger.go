package logger

import (
	"io"
)

func NewTestLogger() Logger {
	return NewLoggerWithWriter(false, io.Discard)
}
