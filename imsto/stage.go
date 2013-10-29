package main

import (
	"fmt"
	// "calf/storage"
	// "calf/image"
	// "os"
)

var cmdStage = &Command{
	UsageLine: "stage [filename] [destname]",
	Short:     "import data from imsto old version or file",
	Long: `
import from a image file
`,
}

func init() {
	cmdStage.Run = runStage
}

func runStage(args []string) bool {
	fmt.Println(cmdStage.Name())
	return false
}
