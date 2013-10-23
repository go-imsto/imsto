package main

import (
	"fmt"
	// "imsto"
	"calf/image"
	"os"
)

var cmdOptimize = &Command{
	UsageLine: "optimize [filename] [destname]",
	Short:     "import data from imsto old version or file",
	Long: `
import from a image file
`,
}

func init() {
	cmdOptimize.Run = runOptimize
}

func runOptimize(args []string) bool {

	if len(args) < 1 {
		//fmt.Println("nothing")
		usage(1)
	}

	file, err := os.Open(args[0])
	defer file.Close()

	if err != nil {
		fmt.Println(err)
		return false
	}

	// write
	dest, err := os.Create(args[1])
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer dest.Close()
	image.RewriteJpeg(file, dest, &image.WriteOption{Quality: 75, StripAll: true})

	return true
}
