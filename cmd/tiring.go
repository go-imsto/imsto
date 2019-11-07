package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/web"
)

var cmdTiring = &Command{
	UsageLine: "tiring -l :8967",
	Short:     "serve tiring http service",
	Long: `
serve tiring http service
`,
}

var (
	maddr string
)

func init() {
	cmdTiring.Run = runTiring
	cmdTiring.Flag.StringVar(&maddr, "l", config.Current.TiringListen, "tcp listen addr")
}

func runTiring(args []string) bool {

	str := fmt.Sprintf("Start Tiring service %s at addr %s", config.Version, maddr)
	fmt.Println(str)
	log.Print(str)
	srv := &http.Server{
		Addr:        maddr,
		Handler:     web.Handler(),
		ReadTimeout: config.Current.ReadTimeout,
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Printf("Fail to start: %s\n", err)
		return false
	}

	return true
}
