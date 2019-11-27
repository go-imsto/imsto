package view

import (
	"html/template"
	"net/http"

	zlog "github.com/go-imsto/imsto/log"
)

func logger() zlog.Logger {
	return zlog.Get()
}

// RenderHTML ...
func RenderHTML(name string, data interface{}, w http.ResponseWriter) (err error) {
	var blob []byte
	blob, err = Asset(name)
	if err != nil {
		logger().Warnw("load fail", "name", name, "err", err)
		return
	}
	var tpl *template.Template
	tpl, err = template.New("default").Parse(string(blob))
	if err != nil {
		logger().Warnw("parse template fail", "err", err)
		return
	}

	err = tpl.Execute(w, data)
	return
}
