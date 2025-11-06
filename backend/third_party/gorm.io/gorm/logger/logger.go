package logger

import (
	"context"
	"time"
)

type LogLevel int

const (
	Silent LogLevel = iota
	Error
	Warn
	Info
)

type Interface interface {
	LogMode(LogLevel) Interface
	Info(context.Context, string, ...interface{})
	Warn(context.Context, string, ...interface{})
	Error(context.Context, string, ...interface{})
	Trace(context.Context, time.Time, func() (string, int64), error)
}
