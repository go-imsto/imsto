package web

import (
	zlog "github.com/go-imsto/imsto/log"
)

func logger() zlog.Logger {
	return zlog.Get()
}
