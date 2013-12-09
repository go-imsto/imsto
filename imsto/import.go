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
	roof           string
	idir           string
	match          string
	arch           string
	include_parent bool
)

func init() {
	cmdImport.Run = runImport
	cmdImport.Flag.StringVar(&roof, "s", "", "config section name")
	cmdImport.Flag.StringVar(&idir, "dir", "", "Import the whole folder recursively if specified.")
	cmdImport.Flag.StringVar(&arch, "archive", "", "Import the whole files in archive.zip if specified.")
	cmdImport.Flag.StringVar(&match, "match", "*.jpg", "pattens of files to import, e.g., *.jpg, *.png, works together with -dir")
	cmdImport.Flag.BoolVar(&include_parent, "iip", false, "is include parent dir name?")
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

func _store_zip(zipfile string) bool {
	log.Printf("reading zip %s", zipfile)
	r, err := zip.OpenReader(zipfile)
	if err != nil {
		fmt.Printf("fail %q\n", err)
		return false
	}
	defer r.Close()

	if len(r.File) == 0 {
		fmt.Printf("fail %q is empty\n", zipfile)
		return false
	}

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

		log.Printf("file: %s\n", f.Name)

		rc, err := f.Open()
		if err != nil {
			log.Print(err)
			continue
		}
		name := _shrink_name(f.Name)

		entry, err := storage.StoredReader(rc, name, roof, uint64(f.FileInfo().ModTime().Unix()))

		_out_entry(entry, name, err)
		rc.Close()

	}
	return true
}

func _store_dir(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// fmt.Printf("path: %s\n", path)
		if err == nil {
			if !info.IsDir() {
				if match != "" {
					if ok, _ := filepath.Match(match, filepath.Base(path)); !ok {
						log.Printf("file %s not match %s", path, match)
						return nil
					}
				}

				_store_file(path, roof)
				return nil
			}
		} else {
			log.Printf("dir walk error: %s", err)
		}
		return err
	})
}

func _store_file(file, roof string) {
	var name string
	if include_parent {
		name = _shrink_name(file)
	} else {
		name = filepath.Base(file)
	}

	// fmt.Printf("%s\n", name)
	entry, err := storage.StoredFile(file, name, roof)
	if err != nil {
		log.Printf("store file error: %s", err)
	}
	_out_entry(entry, name, err)
}

func _out_entry(entry *storage.Entry, name string, err error) {
	if err != nil {
		fmt.Printf("fail %q %q\n", name, err)
	} else {
		fmt.Printf("ok %s %s %q\n", entry.Id, entry.Path, name)
	}

}

func _shrink_name(fname string) string {
	a := strings.Split(fname, "/")
	l := len(a)
	if l > 1 {
		return strings.Join(a[l-2:l], "/")
	}
	return fname

}
