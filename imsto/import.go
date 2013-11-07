package main

import (
	// "calf/image"
	"calf/storage"
	// "encoding/json"
	"fmt"
	"io/ioutil"
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

	var err error

	if _, err = os.Stat(args[0]); err != nil {
		if os.IsNotExist(err) {
			fmt.Println(err)
			return false
		}
	}
	if err != nil {
		fmt.Println(err)
		return false
	}
	name := path.Base(args[0])

	var (
		data  []byte
		entry *storage.Entry
	)

	data, err = ioutil.ReadFile(args[0])

	entry, err = storage.NewEntry(data)

	if err != nil {
		fmt.Println(err)
		return false
	}
	// fmt.Println(entry)
	fmt.Printf("new id: %v, size: %d, path: %v\n", entry.Id, entry.Size, entry.Path)

	// var b []byte
	// b, err = json.Marshal(entry)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return false
	// }

	// fmt.Println("json:", string(b))

	entry.Name = name

	var mw storage.MetaWrapper
	mw = storage.NewMetaWrapper("")
	// fmt.Println("mw", mw)

	err = mw.Store(entry)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}
