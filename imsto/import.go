package main

import (
	// "wpst.me/calf/image"
	"wpst.me/calf/storage"
	// "encoding/json"
	"fmt"
	// "io/ioutil"
	// "os"
	// "path"
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

	section := ""
	var entry *storage.Entry
	entry, err := storage.StoredFile(args[0], section)

	if err != nil {
		fmt.Printf("fail %s\n", err)
		return false
	}

	fmt.Printf("ok %s %s\n", entry.Id, entry.Path)

	// var mw storage.MetaWrapper
	// mw = storage.NewMetaWrapper("")
	// // fmt.Println("mw", mw)

	// err = mw.Store(entry)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return false
	// }

	return true
}
