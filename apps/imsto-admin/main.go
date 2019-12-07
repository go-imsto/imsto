package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/bmizerany/pat"
	staffio "github.com/liut/staffio-client"
	"go.uber.org/zap"

	"github.com/go-imsto/imsto/config"
	zlog "github.com/go-imsto/imsto/log"
	"github.com/go-imsto/imsto/web/admin/static"
	"github.com/go-imsto/imsto/web/admin/view"
)

var (
	apiKey string
)

func main() {
	var addr string
	flag.StringVar(&addr, "l", ":8970", "listen addr")

	var zlogger *zap.Logger
	if config.InDevelop() {
		zlogger, _ = zap.NewDevelopment()
		zlogger.Debug("logger start")
	} else {
		zlogger, _ = zap.NewProduction()
	}
	defer zlogger.Sync() // flushes buffer, if any
	sugar := zlogger.Sugar()

	zlog.Set(sugar)

	apiKey = config.EnvOr("IMSTO_DEMO_API_KEY", "")

	mux := pat.New()
	loginPath := "/auth/login"
	staffio.SetLoginPath(loginPath)
	staffio.SetAdminPath("/")

	mux.Get(loginPath, http.HandlerFunc(staffio.LoginHandler))
	mux.Get("/auth/callback", staffio.AuthCodeCallback("verandah"))
	authF1 := staffio.AuthMiddleware(true) // auto redirect

	mux.Get("/", authF1(http.HandlerFunc(handleIndex)))
	mux.Get("/gallery", httpLogger(authF1(http.HandlerFunc(handlerGallery))))
	mux.Get("/upload", httpLogger(authF1(http.HandlerFunc(handlerUpload))))

	mux.Get("/static/", http.StripPrefix("/static/", static.Server))

	logger().Infow("listen admin", "addr", addr)
	logger().Fatalw("listen fail", "err", http.ListenAndServe(":8970", mux))

}

func logger() zlog.Logger {
	return zlog.Get()
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	view.Render("index.htm", nil, w)
}

func handlerGallery(w http.ResponseWriter, r *http.Request) {
	view.Render("gallery.htm", map[string]string{"APIKey": apiKey}, w)
}

func handlerUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xhtml+xml")
	view.Render("upload.htm", map[string]string{"APIKey": apiKey}, w)
}

func httpLogger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		t1 := time.Now()
		defer func() {
			logger().Debugw("request", "method", r.Method, "uri", r.RequestURI, "remote", r.RemoteAddr, "agent", r.UserAgent(), "ts", time.Since(t1))
		}()

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
