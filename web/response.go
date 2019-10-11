package web

import (
	"encoding/json"
	"fmt"
	"net"
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
func writeJsonQuiet(w http.ResponseWriter, r *http.Request, obj interface{}) {
	if err := writeJson(w, r, obj); err != nil {
		logger().Warnw("error writing JSON %s: %s", obj, err.Error())
	}
}

func writeJsonError(w http.ResponseWriter, r *http.Request, err error) {
	if r.Method == "GET" || r.Method == "HEAD" {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
		w.Header().Set("Pragma", "no-cache")
	}

	res := newApiRes(newApiMeta(false), nil)
	res["error"] = newApiError(err)

	writeJsonQuiet(w, r, res)
}

func secure(whiteList []string, f http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(whiteList) == 0 {
			f(w, r)
			return
		}
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			for _, ip := range whiteList {
				if ip == host {
					f(w, r)
					return
				}
			}
		}
		w.WriteHeader(http.StatusForbidden)
		writeJsonQuiet(w, r, map[string]interface{}{"error": "No write permisson from " + host})
	})
}
