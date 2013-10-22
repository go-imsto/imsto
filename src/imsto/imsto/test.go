package main

import (
	"bufio"
	"fmt"
	"imsto"
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

const _head_size = 8

func init() {
	cmdTest.Run = testApp
}

func testApp(args []string) bool {
	al := len(args)
	if al == 0 {
		fmt.Println("nothing")
	} else {
		fmt.Println(args[0])
	}

	if al > 1 && args[0] == "imagetype" {

		data, err := readImageHead(args[1])

		if err != nil {
			fmt.Println(err)
			return false
		}

		t := imsto.GuessImageType(&data)

		fmt.Println(t)
		fmt.Println(imsto.ExtByImageType(t))
	} else if al > 1 && args[0] == "mimetype" {

		ext := path.Ext(args[1])
		fmt.Println(ext)
		mimetype := mime.TypeByExtension(ext)
		fmt.Println(mimetype)
	}

	return true
}

func readImageHead(fname string) ([]byte, error) {

	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}

	r := bufio.NewReaderSize(file, _head_size)
	var data []byte

	if data, err = r.Peek(_head_size); err != nil {
		return nil, err
	}

	if err = file.Close(); err != nil {
		return nil, err
	}
	// fmt.Println(data)
	return data, nil
}
