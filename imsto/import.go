package main

import (
	"archive/zip"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"wpst.me/calf/storage"
)

const import_usage = `import -s roof file1 [file2] [file3]
	import -s roof -dir directory
	import -s roof -archive archive.zip
`

var cmdImport = &Command{
	UsageLine: import_usage,
	Short:     "import data from local file",
	Long: `
import from local system
`,
}

var (
	roof  string
	idir  string
	match string
	arch  string
)

func init() {
	cmdImport.Run = runImport
	cmdImport.Flag.StringVar(&roof, "s", "", "config section name")
	cmdImport.Flag.StringVar(&idir, "dir", "", "Import the whole folder recursively if specified.")
	cmdImport.Flag.StringVar(&arch, "archive", "", "Import the whole files in archive.zip if specified.")
	cmdImport.Flag.StringVar(&match, "match", "*.jpg", "pattens of files to import, e.g., *.jpg, *.png, works together with -dir")
}

func runImport(args []string) bool {

	if roof == "" {
		return false
	}

	if len(args) == 0 {
		// if idir == "" && arch == "" {
		// 	return false
		// }
		if idir != "" {
			_store_dir(idir)
		} else if arch != "" {
			_store_zip(arch)
		} else {
			return false
		}

	} else {
		for _, file := range args {
			_store_file(file, roof)
		}
	}

	return true
}

func _store_zip(zipfile string) {
	log.Printf("reading zip %s", zipfile)
	r, err := zip.OpenReader(zipfile)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			log.Printf("dir %s\n", f.Name)
			// fmt.Printf("fail %s is a dir", f.Name)
			continue
		}

		if ok, _ := filepath.Match(match, filepath.Base(f.Name)); !ok {
			log.Printf("fail '%s' not match '%s'\n", f.Name, match)
			continue
		}

		var name string
		a := strings.Split(f.Name, "/")
		l := len(a)
		if l > 2 {
			name = strings.Join(a[l-2:l], "/")
		} else {
			name = f.Name
		}

		log.Printf("file: %s\n", f.Name)
		rc, err := f.Open()

		entry, err := storage.StoredReader(rc, name, roof, uint64(f.FileInfo().ModTime().Unix()))

		if err != nil {
			fmt.Printf("fail '%s' error:%s\n", name, err)
			// fmt.Printf("entry meta: %s\n", entry.Meta)
		}

		fmt.Printf("ok %s %s '%s'\n", entry.Id, entry.Path, entry.Name)
		rc.Close()

	}
}

func _store_dir(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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
				// TODO: break or not break?
				// return e
				return nil
			}
		} else {
			log.Printf("dir walk error: %s", err)
		}
		return err
	})
}

func _store_file(file, roof string) error {
	entry, err := storage.StoredFile(file, roof)

	if err != nil {
		fmt.Printf("fail '%s' error:%s\n", file, err)
		// fmt.Printf("entry meta: %s\n", entry.Meta)
		return err
	}

	fmt.Printf("ok %s %s '%s'\n", entry.Id, entry.Path, entry.Name)
	return nil
}
