package view

import (
	"html/template"
	"net/http"

	zlog "github.com/go-imsto/imsto/log"
)

func logger() zlog.Logger {
	return zlog.Get()
}

//go:generate staticfiles --package view -o files.go ../templates
//go:generate staticfiles --package static -o ../static/files.go ../../../apps/static

// Render ...
func Render(name string, data interface{}, w http.ResponseWriter) (err error) {
	text := Data(name)
	if len(text) == 0 {
		logger().Infow("load template empty", "name", name)
		return
	}
	var tpl *template.Template
	tpl, err = template.New("default").Parse(text)
	if err != nil {
		logger().Warnw("parse template fail", "err", err)
		return
	}

	err = tpl.Execute(w, data)
	return
}
