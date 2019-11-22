package web

import (
	"net"
	"net/http"

	"github.com/go-imsto/imsto/config"
)

func secure(f http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(config.Current.WhiteList) == 0 {
			f(w, r)
			return
		}
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			ip := net.ParseIP(host)
			for _, ipn := range config.Current.WhiteList {
				if ipn.Contains(ip) {
					f(w, r)
					return
				}
			}
		}
		w.WriteHeader(http.StatusForbidden)
		writeJSONQuiet(w, r, map[string]interface{}{"error": "No write permisson from " + host})
	})
}
