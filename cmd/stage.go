package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/web"
)

var cmdStage = &Command{
	UsageLine: "stage -l 8968",
	Short:     "stage is a image handler",
	Long: `
stage is a image handler.
`,
}

var (
	saddr string
)

func init() {
	cmdStage.Run = runStage
	cmdStage.Flag.StringVar(&saddr, "l", config.Current.StageListen, "tcp listen addr")
}

func runStage(args []string) bool {

	mux := http.NewServeMux()
	mux.HandleFunc("/", web.StageHandler)

	str := fmt.Sprintf("Start Stage service %s at addr %s", config.Version, saddr)
	fmt.Println(str)
	log.Print(str)
	srv := &http.Server{
		Addr:        saddr,
		Handler:     mux,
		ReadTimeout: config.Current.ReadTimeout,
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Printf("Fail to start:%s\n", err)
		return false
	}

	return true
}
