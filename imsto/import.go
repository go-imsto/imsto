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

func runImport(args []string) bool {
	if len(args) == 0 {
		fmt.Println("nothing")
		return false
	} else {
		fmt.Println(args[0])
	}

	if _, err := os.Stat(args[0]); err != nil {
		if os.IsNotExist(err) {
			fmt.Println(err)
			return false
		}
	}

	im, err := image.Open(args[0])

	if err != nil {
		fmt.Println(err)
		return false
	}

	defer im.Close()

	ia := im.GetAttr()

	fmt.Print("ia: ")
	fmt.Println(ia)

	file, err := os.Open(args[0])
	defer file.Close()

	if err != nil {
		fmt.Println(err)
		return false
	}

	var (
		entry *imsto.Entry
	)

	entry, err = imsto.NewEntryByReader(file)

	if err != nil {
		fmt.Println(err)
		return false
	}

	// fmt.Println(entry)
	fmt.Printf("new id: %v, size: %d, path: %v\n", entry.Id, entry.Size, entry.Path)

	return true
}
