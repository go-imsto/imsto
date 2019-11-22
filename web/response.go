package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type apiRes map[string]interface{}
type apiMeta map[string]interface{}
type apiError struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"message,omitempty"`
	err  error
}

func newApiRes(meta apiMeta, data interface{}) apiRes {
	res := make(apiRes)
	res["meta"] = meta
	res["data"] = data
	return res
}

func newApiMeta(ok bool) apiMeta {
	meta := make(apiMeta)
	meta["ok"] = ok
	return meta
}

func newApiError(err error) apiError {
	ae := apiError{err: err}
	ae.Msg = err.Error()
	return ae
}

func writeJson(w http.ResponseWriter, r *http.Request, obj interface{}) (err error) {
	w.Header().Set("Content-Type", "application/json")
	var bytes []byte
	if r.FormValue("pretty") != "" {
		bytes, err = json.MarshalIndent(obj, "", "  ")
	} else {
		bytes, err = json.Marshal(obj)
	}
	if err != nil {
		return
	}
	callback := r.FormValue("callback")
	if callback == "" {
		_, err = w.Write(bytes)
	} else {
		if _, err = w.Write([]uint8(callback)); err != nil {
			return
		}
		if _, err = w.Write([]uint8("(")); err != nil {
			return
		}
		fmt.Fprint(w, string(bytes))
		if _, err = w.Write([]uint8(")")); err != nil {
			return
		}
	}
	return
}

// wrapper for writeJson - just logs errors
func writeJSONQuiet(w http.ResponseWriter, r *http.Request, obj interface{}) {
	if err := writeJson(w, r, obj); err != nil {
		logger().Warnw("error writing JSON %s: %s", obj, err.Error())
	}
}

func writeJSONError(w http.ResponseWriter, r *http.Request, err error) {

	if r.Method == "GET" || r.Method == "HEAD" {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
		w.Header().Set("Pragma", "no-cache")
	}

	res := newApiRes(newApiMeta(false), nil)
	res["error"] = newApiError(err)

	writeJSONQuiet(w, r, res)
}
