package main

import (
	// "fmt"
	"log"
	// "calf/storage"
	// "calf/image"
	// "os"
	// "github.com/ugorji/go/codec"
	// "net"
	"net/http"
	"strconv"
	"time"
)

var cmdTiring = &Command{
	UsageLine: "tiring -port 5564",
	Short:     "serve tiring tcp service",
	Long: `
serve tiring tcp service
`,
}

var (
	mport           = cmdTiring.Flag.Int("port", 5564, "tcp listen port")
	mReadTimeout    = cmdTiring.Flag.Int("readTimeout", 3, "connection read timeout in seconds")
	whiteListOption = cmdTiring.Flag.String("whiteList", "", "comma separated Ip addresses having write permission. No limit if empty.")
	whiteList       []string
)

func init() {
	cmdTiring.Run = runTiring
	cmdTiring.IsDebug = cmdTiring.Flag.Bool("debug", false, "enable debug mode")
}

func browseHandler(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	m["Version"] = VERSION
	writeJsonQuiet(w, r, m)
}

func runTiring(args []string) bool {
	// fmt.Println(cmdTiring.Name())
	var e error
	http.HandleFunc("/", browseHandler)

	log.Print("Start Tiring service", VERSION, "at port", strconv.Itoa(*mport))
	srv := &http.Server{
		Addr:        ":" + strconv.Itoa(*mport),
		Handler:     http.DefaultServeMux,
		ReadTimeout: time.Duration(*mReadTimeout) * time.Second,
	}
	e = srv.ListenAndServe()
	if e != nil {
		log.Printf("Fail to start:%s\n", e)
	}

	return true
}
