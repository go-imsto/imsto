package main

import (
	"bufio"
	cimg "calf/image"
	"calf/storage"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime"
	"os"
	"path"
)

var cmdTest = &Command{
	UsageLine: "test",
	Short:     "run all tests from the command-line",
	Long: `
Just a test command
`,
}

var (
	url = cmdTest.Flag.String("path", "", "entry path for load")
)

const _head_size = 8

func init() {
	cmdTest.Run = testApp
}

func testApp(args []string) bool {
	if *url != "" {
		fmt.Println("url: ", *url)
		section := ""
		item, err := storage.LoadPath(*url, section)
		if err != nil {
			fmt.Println("Err: ", err)
			return false
		}
		fmt.Print(item)
	}

	al := len(args)
	fmt.Println(args)
	if al == 0 {
		fmt.Println("nothing")
		return false
	}

	if al > 1 && args[0] == "imagetype" {

		file, err := os.Open(args[1])

		if err != nil {
			fmt.Println(err)
			return false
		}
		defer file.Close()
		var (
			t   cimg.TypeId
			ext string
		)
		t, ext, err = cimg.GuessType(file)

		fmt.Println(t, ext)
	} else if al > 1 && args[0] == "mimetype" {

		ext := path.Ext(args[1])
		fmt.Println(ext)
		mimetype := mime.TypeByExtension(ext)
		fmt.Println(mimetype)
	} else if al > 1 && args[0] == "image" {
		im, format, err := readImage(args[1])
		if err != nil {
			fmt.Println(err)
			return false
		}
		fmt.Println(im.Bounds(), format, err)

		if al > 2 {
			var outfile *os.File
			outfile, err = os.Create(args[2])
			if err != nil {
				fmt.Println(err)
				return false
			}
			if format == "png" {
				err = png.Encode(outfile, im)
			} else if format == "jpeg" {
				err = jpeg.Encode(outfile, im, &jpeg.Options{Quality: 75})
			} else {
				fmt.Println("unsupported format")
				return false
			}
			if err != nil {
				fmt.Println(err)
				return false
			}
		}
	}

	return true
}

func readImage(filename string) (image.Image, string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	return image.Decode(bufio.NewReader(f))
}

func readImageConfig(filename string) (image.Config, string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return image.Config{}, "", err
	}
	defer f.Close()
	return image.DecodeConfig(bufio.NewReader(f))
}
