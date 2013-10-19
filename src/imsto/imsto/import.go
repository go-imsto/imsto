package main

import (
	"fmt"
	"imsto"
	"imsto/image"
	"os"
)

var cmdImport = &Command{
	UsageLine: "import [filename]",
	Short:     "import data from imsto old version or file",
	Long: `
import from a image file
`,
}

func init() {
	cmdImport.Run = runImport
}

func runImport(args []string) {
	if len(args) == 0 {
		fmt.Println("nothing")
		return
	} else {
		fmt.Println(args[0])
	}

	ia, _ := image.ReadJpeg(args[0])

	// ia := image.GetImageAttr(args[0])

	fmt.Println(ia)

	file, err := os.Open(args[0])
	defer file.Close()

	if err != nil {
		fmt.Println(err)
	}

	var (
		entry *imsto.Entry
	)

	entry, err = imsto.NewEntryFromIo(file)

	if err != nil {
		fmt.Println(err)
	}

	// fmt.Println(entry)
	fmt.Printf("new id: %v, size: %d, path: %v\n", entry.Id, entry.Size, entry.Path)

}
