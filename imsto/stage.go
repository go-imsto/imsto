package main

import (
	"fmt"
	// "calf/storage"
	// "calf/image"
	// "os"
)

var cmdStage = &Command{
	UsageLine: "stage -port 5580",
	Short:     "stage is a image handler",
	Long: `
stage is a image handler.
`,
}

var (
	sport = cmdStage.Flag.Int("port", 5580, "tcp listen port")
)

func init() {
	cmdStage.Run = runStage
}

func runStage(args []string) bool {
	// TODO:
	fmt.Println(cmdStage.Name())
	return false
}
