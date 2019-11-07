package cmd

import (
	"fmt"

	"github.com/go-imsto/imsto/config"
)

var cmdBundle = &Command{
	UsageLine: "bundle",
	Short:     "run all services",
	Long:      ``,
}

func init() {
	cmdBundle.Run = runBundle
}

func runBundle(args []string) bool {
	fmt.Printf("Start RPC/Stage/Tiring service %s\n", config.Version)
	go runTiring(args)
	go runStage(args)
	return runRPC(args)
}
