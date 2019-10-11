package cmd

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/web"
)

var cmdTiring = &Command{
	UsageLine: "tiring -port 8964 -whiteList=\"127.0.0.1\"",
	Short:     "serve tiring http service",
	Long: `
serve tiring http service
`,
}

var (
	mport           = cmdTiring.Flag.Int("port", 8964, "tcp listen port")
	mReadTimeout    = cmdTiring.Flag.Int("readTimeout", 3, "connection read timeout in seconds")
	mMaxCpu         = cmdTiring.Flag.Int("maxCpu", 0, "maximum number of CPUs. 0 means all available CPUs")
	whiteListOption = cmdTiring.Flag.String("whiteList", "", "comma separated Ip addresses having write permission. No limit if empty.")
	whiteList       []string
)

func init() {
	cmdTiring.Run = runTiring
}

func runTiring(args []string) bool {
	if *mMaxCpu < 1 {
		*mMaxCpu = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(*mMaxCpu)

	if *whiteListOption != "" {
		whiteList = strings.Split(*whiteListOption, ",")
	}

	str := fmt.Sprintf("Start Tiring service %s at port %d", config.Version, *mport)
	fmt.Println(str)
	log.Print(str)
	srv := &http.Server{
		Addr:        ":" + strconv.Itoa(*mport),
		Handler:     web.Handler(),
		ReadTimeout: time.Duration(*mReadTimeout) * time.Second,
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Printf("Fail to start: %s\n", err)
		return false
	}

	return true
}
