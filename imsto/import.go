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
	UsageLine: "import -s roof [filename]",
	Short:     "import data from local file",
	Long: `
import from a image file
`,
}

var (
	roof string
)

func init() {
	cmdImport.Run = runImport
	cmdImport.Flag.StringVar(&roof, "s", "", "config section name")
}

func runImport(args []string) bool {

	if roof == "" || len(args) == 0 {
		// fmt.Println("nothing")
		return false
	}

	var entry *storage.Entry
	entry, err := storage.StoredFile(args[0], roof)

	if err != nil {
		fmt.Printf("fail %s\n", err)
		return false
	}

	fmt.Printf("ok %s %s\n", entry.Id, entry.Path)

	return true
}
