package main

import (
	"flag"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"image/png"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
)

var (
	addr string
	size int
)

type httpServer struct{}

func (s httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	str := r.FormValue("c")
	if str == "" {
		log.Print("empty content")
		return
	}

	w.Header().Set("Content-Type", "image/png")

	qrcode, err := qr.Encode(str, qr.L, qr.Auto)
	if err != nil {
		log.Println(err)
	} else {
		qrcode, err = barcode.Scale(qrcode, size, size)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("gen barcode: %s", str)
			png.Encode(w, qrcode)
		}
	}
}

func init() {
	flag.StringVar(&addr, "addr", "127.0.0.1:9001", "listen address")
	flag.IntVar(&size, "size", 110, "barcode dimension")
}

func main() {
	var (
		l   net.Listener
		err error
	)
	flag.Parse()

	if size < 100 {
		size = 100
	} else if size > 720 {
		size = 720
	}

	if addr[0] == '/' {
		l, err = net.Listen("unix", addr)
	} else {
		l, err = net.Listen("tcp", addr)
	}

	if err != nil {
		log.Println(err)
	}

	srv := new(httpServer)
	fcgi.Serve(l, srv)
}
