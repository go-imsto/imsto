package cmd

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/web"
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

func runStage(args []string) bool {
	if *sMaxCpu < 1 {
		*sMaxCpu = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(*sMaxCpu)

	var e error
	http.HandleFunc("/", web.StageHandler)

	str := fmt.Sprintf("Start Stage service %s at port %d", config.Version, *sport)
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
