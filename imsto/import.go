package main

import (
	// "calf/image"
	"calf/storage"
	"fmt"
	"os"
	"path"
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

	var (
		err  error
		file *os.File
		// im   image.Image
	)
	// file, err = os.Open(args[0])

	// if err != nil {
	// 	fmt.Println(err)
	// 	return false
	// }

	// im, err = image.Open(file)

	// if err != nil {
	// 	fmt.Println(err)
	// 	return false
	// }

	// file.Close()
	// defer im.Close()

	// ia := im.GetAttr()

	// fmt.Print("ia: ")
	// fmt.Println(ia)

	file, err = os.Open(args[0])
	defer file.Close()

	if err != nil {
		fmt.Println(err)
		return false
	}
	name := path.Base(args[0])

	var (
		entry *storage.Entry
	)

	entry, err = storage.NewEntry(file)

	if err != nil {
		fmt.Println(err)
		return false
	}

	entry.Name = name

	// fmt.Println(entry)
	fmt.Printf("new id: %v, size: %d, path: %v\n", entry.Id, entry.Size, entry.Path)

	var mw storage.MetaWrapper
	mw = storage.NewMetaWrapper("")

	mw.Store(entry)
	fmt.Println(mw)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}
