package web

import (
	"net/http"

	"github.com/go-playground/form"
)

var (
	formDecoder = form.NewDecoder()
)

// Bind ...
func Bind(req *http.Request, obj interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	req.ParseMultipartForm(32 << 10) // 32 MB
	if err := formDecoder.Decode(obj, req.Form); err != nil {
		return err
	}
	return nil
}

type uploadSchema struct {
	APIKey string `form:"api_key"`
	Token  string `form:"token"`
	Roof   string `form:"roof"`
	User   int    `form:"user"`
	Tags   string `form:"tags"`
}

type tokenSchema struct {
	APIKey string `form:"api_key"`
	Roof   string `form:"roof"`
	User   int    `form:"user"`
}

type ticketSchema struct {
	APIKey string `form:"api_key"`
	Token  string `form:"token"`
	Roof   string `form:"roof"`
	User   int    `form:"user"`
	Prompt string `form:"prompt"`
}
