package web

import (
	"net"
	"net/http"
	"os"
	"strings"
)

var (
	whiteList = []string{}
)

func init() {
	if str, ok := os.LookupEnv("IMSTO_WHITE_LIST"); ok && len(str) > 0 {
		whiteList = strings.Split(str, ",")
	}
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
