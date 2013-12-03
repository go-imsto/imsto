package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"wpst.me/calf/storage"
)

var cmdImport = &Command{
	UsageLine: "import -s roof file1 [file2] [file3]\n\t\timport -s roof -dir directory",
	Short:     "import data from local file",
	Long: `
import from a image file
`,
}

var (
	roof  string
	idir  string
	match string
)

func init() {
	cmdImport.Run = runImport
	cmdImport.Flag.StringVar(&roof, "s", "", "config section name")
	cmdImport.Flag.StringVar(&idir, "dir", "", "Import the whole folder recursively if specified.")
	cmdImport.Flag.StringVar(&match, "match", "*.jpg", "pattens of files to import, e.g., *.jpg, *.png, works together with -dir")
}

func runImport(args []string) bool {

	if roof == "" {
		return false
	}

	if len(args) == 0 {
		if idir == "" {
			return false
		}
		filepath.Walk(idir, func(path string, info os.FileInfo, err error) error {
			// fmt.Printf("path: %s\n", path)
			if err == nil {
				if !info.IsDir() {
					if match != "" {
						if ok, _ := filepath.Match(match, filepath.Base(path)); !ok {
							return nil
						}
					}
					// return nil
					e := _store_file(path, roof)
					if e != nil {
						log.Printf("store file error: %s", e)
					}
					// TODO: to be or not to be?
					// return e
					return nil
				}
			} else {
				log.Printf("dir walk error: %s", err)
			}
			return err
		})

	} else {
		for _, file := range args {
			_store_file(file, roof)
		}
	}

	return true
}

func _store_file(file, roof string) error {
	var entry *storage.Entry
	entry, err := storage.StoredFile(file, roof)

	if err != nil {
		fmt.Printf("fail '%s' error:%s\n", file, err)
		// fmt.Printf("entry meta: %s\n", entry.Meta)
		return err
	}

	fmt.Printf("ok %s %s '%s'\n", entry.Id, entry.Path, entry.Name)
	return nil
}
