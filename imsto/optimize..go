package main

import (
	"fmt"
	// "imsto"
	"calf/image"
	"os"
)

var cmdOptimize = &Command{
	UsageLine: "optimize [filename] [destname]",
	Short:     "optimize a jpeg file",
	Long: `
optimize a jpeg file
`,
}

var (
	quality = cmdOptimize.Flag.Int("q", 88, "max quality")
	strip   = cmdOptimize.Flag.String("s", "all", "strip [all]")
)

func init() {
	cmdOptimize.Run = runOptimize
}

func runOptimize(args []string) bool {

	var (
		im image.Image
	)
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

	im, err = image.Open(file)

	if err != nil {
		fmt.Println(err)
		return false
	}

	defer im.Close()
	wopt := image.WriteOption{Quality: 88, StripAll: true}
	if *quality > 60 && *quality < 100 {
		wopt.Quality = image.Quality(*quality)
	}
	im.SetOption(wopt)

	// write
	dest, err := os.Create(args[1])
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer dest.Close()
	// image.OptimizeJpeg(file, dest, &image.WriteOption{Quality: 75, StripAll: true})

	err = im.WriteTo(dest)

	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}
