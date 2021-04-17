package util

import "io"

type Logger interface {
	io.Writer
	Infof(msg string, args ...interface{})
	Info(args ...interface{})
	Errorf(msg string, args ...interface{})
	Error(args ...interface{})
}
