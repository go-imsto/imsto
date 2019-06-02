package cmd

import (
	"fmt"
	"github.com/go-imsto/imsto/storage"
	"io"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

var cmdStage = &Command{
	UsageLine: "stage -port 8968",
	Short:     "stage is a image handler",
	Long: `
stage is a image handler.
`,
}

var (
	sport        = cmdStage.Flag.Int("port", 8968, "tcp listen port")
	sReadTimeout = cmdStage.Flag.Int("readTimeout", 15, "connection read timeout in seconds")
	sMaxCpu      = cmdStage.Flag.Int("maxCpu", 0, "maximum number of CPUs. 0 means all available CPUs")
)

func init() {
	cmdStage.Run = runStage
}

func StageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Server", "IMSTO STAGE")

	item, err := storage.LoadPath(r.URL.Path)

	if err != nil {
		logger().Warnw("loadPath fail", "ref", r.Referer(), "err", err)
		switch err.(type) {
		case *storage.HttpError:
			ie := err.(*storage.HttpError)
			if ie.Code == 302 {
				// log.Printf("redirect to %s", ie.Path)
				http.Redirect(w, r, ie.Path, ie.Code)
				return
			}
			// w.WriteHeader(ie.Code)
			http.Error(w, ie.Text, ie.Code)
			return
		}
		writeJsonError(w, r, err)

		return
	}

	// log.Print(item)

	c := func(file io.ReadSeeker) {
		http.ServeContent(w, r, item.Name(), item.Modified(), file)
	}
	err = item.Walk(c)
	if err != nil {
		logger().Warnw("item walk fail", "item", item, "err", err)
		writeJsonError(w, r, err)
		return
	}
}

func runStage(args []string) bool {
	if *sMaxCpu < 1 {
		*sMaxCpu = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(*sMaxCpu)

	var e error
	http.HandleFunc("/", StageHandler)

	// log.Print("Start Stage service ", VERSION, " at port ", strconv.Itoa(*sport))
	str := fmt.Sprintf("Start Stage service %s at port %d", VERSION, *sport)
	fmt.Println(str)
	log.Print(str)
	srv := &http.Server{
		Addr:        ":" + strconv.Itoa(*sport),
		Handler:     http.DefaultServeMux,
		ReadTimeout: time.Duration(*sReadTimeout) * time.Second,
	}
	e = srv.ListenAndServe()
	if e != nil {
		log.Printf("Fail to start:%s\n", e)
		return false
	}

	return true
}
