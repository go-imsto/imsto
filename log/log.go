package log

import (
	syslog "log"
	"log/slog"
)

type logger struct{}

// Default 默认实例
var Default Logger

func init() {
	syslog.SetFlags(syslog.Ltime | syslog.Lshortfile)
	Default = &logger{}
}

func Set(logger Logger) {
	if logger != nil {
		Default = logger
	}
}

func Get() Logger {
	return Default
}

func (z *logger) Debugw(msg string, keysAndValues ...any) {
	slog.Debug(msg, keysAndValues...)
}

func (z *logger) Infow(msg string, keysAndValues ...any) {
	slog.Info(msg, keysAndValues...)
}

func (z *logger) Warnw(msg string, keysAndValues ...any) {
	slog.Warn(msg, keysAndValues...)
}

func (z *logger) Errorw(msg string, keysAndValues ...any) {
	slog.Error(msg, keysAndValues...)
}

func (z *logger) Fatalw(msg string, keysAndValues ...any) {
	syslog.Fatal(msg, keysAndValues)
}

func Debugw(msg string, keysAndValues ...any) {
	slog.Debug(msg, keysAndValues...)
}

func Infow(msg string, keysAndValues ...any) {
	slog.Info(msg, keysAndValues...)
}

func Warnw(msg string, keysAndValues ...any) {
	slog.Warn(msg, keysAndValues...)
}

func Errorw(msg string, keysAndValues ...any) {
	slog.Error(msg, keysAndValues...)
}

func Fatalw(msg string, keysAndValues ...any) {
	Default.Fatalw(msg, keysAndValues)
}
