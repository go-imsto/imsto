package cmd

import (
	"fmt"
	"log"
	"net/http"
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
)

func init() {
	cmdStage.Run = runStage
}

func runStage(args []string) bool {

	mux := http.NewServeMux()
	mux.HandleFunc("/", web.StageHandler)

	str := fmt.Sprintf("Start Stage service %s at port %d", config.Version, *sport)
	fmt.Println(str)
	log.Print(str)
	srv := &http.Server{
		Addr:        ":" + strconv.Itoa(*sport),
		Handler:     mux,
		ReadTimeout: time.Duration(*sReadTimeout) * time.Second,
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Printf("Fail to start:%s\n", err)
		return false
	}

	return true
}
