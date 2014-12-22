package main

import (
	"flag"
	"fmt"
	"os"
	"wpst.me/calf/image"
)

const usage_line = `optimize [filename] [destname]
`

const short_desc = "optimize a jpeg file"

var (
	quality int
	strip   string
)

func usage() {
	fmt.Printf("Usage: \t%s\nDefault Usage:\n", usage_line)
	flag.PrintDefaults()
	fmt.Println("\nDescription:\n   " + short_desc + "\n")
}

func init() {
	flag.IntVar(&quality, "q", 88, "max quality")
	flag.StringVar(&strip, "s", "all", "strip [all]")
	flag.Parse()
}

func main() {

	var (
		im image.Image
	)
	args := flag.Args()
	if len(args) < 1 {
		//fmt.Println("nothing")
		usage()
		return
	}

	file, err := os.Open(args[0])
	defer file.Close()

	if err != nil {
		fmt.Println(err)
		return
	}

	im, err = image.Open(file)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer im.Close()
	attr := im.GetAttr()
	org_quality := int(attr.Quality)
	fmt.Printf("org quality: %d\n", org_quality)
	if quality < org_quality {
		wopt := image.WriteOption{Quality: image.Quality(quality), StripAll: true}
		if quality > 60 && quality < 100 {
			wopt.Quality = image.Quality(quality)
		}
		im.SetOption(wopt)
	}

	// write
	dest, err := os.Create(args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dest.Close()
	// image.OptimizeJpeg(file, dest, &image.WriteOption{Quality: 75, StripAll: true})

	err = im.WriteTo(dest)

	if err != nil {
		fmt.Println(err)
		return
	}

	return
}
