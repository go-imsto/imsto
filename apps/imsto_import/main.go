package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/storage"
)

const usage_line = `import -s roof [-tag=foo,bar] [-author=aid] file1 [file2] [file3]
	import -s roof -dir directory [-dt] [-author=aid]
	import -s roof -archive archive.zip [-author=aid]
	import -s roof -ready
`

const short_desc = "import data from local file"

var (
	cfgDir         string
	roof           string
	idir           string
	match          string
	arch           string
	include_parent bool
	readydone      bool
	dirAsTag       bool
	tags           string
	author         int
)

func usage() {
	fmt.Printf("Usage: \t%s\nDefault Usage:\n", usage_line)
	flag.PrintDefaults()
	fmt.Println("\nDescription:\n   " + short_desc + "\n")
}

func init() {
	flag.StringVar(&cfgDir, "conf", "/etc/imsto", "app conf dir")
	flag.StringVar(&roof, "s", "", "config section name")
	flag.StringVar(&arch, "archive", "", "Import the whole files in archive.zip if specified.")
	flag.StringVar(&match, "match", "*.jpg", "pattens of files to import, e.g., *.jpg, *.png, works together with -dir")
	flag.BoolVar(&include_parent, "iip", false, "is include parent dir name?")
	flag.BoolVar(&dirAsTag, "dt", false, "check file's directory as tag[s]")
	flag.StringVar(&tags, "tag", "", "give one or more tags")
	flag.IntVar(&author, "author", 0, "give a author_id")

	flag.Parse()

}

func main() {
	if roof == "" {
		usage()
		return
	}

	if !config.HasSection(roof) {
		fmt.Printf("roof [%s] not found\n", roof)
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		// if idir == "" && arch == "" {
		// 	return false
		// }
		if idir != "" {
			_store_dir(idir)
		} else if arch != "" {
			_store_zip(arch)
		} else {
			usage()
			return
		}

	} else {
		for _, file := range args {
			_store_file(file, roof, tags)
		}
	}

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

		var buf []byte
		buf, err = ioutil.ReadAll(rc)
		if err != nil {
			log.Print(err)
			return false
		}
		rc.Close()
		entry, err := storage.PrepareReader(bytes.NewReader(buf), name)
		if err != nil {
			log.Print(err)
			continue
		}

		if author > 0 {
			entry.Author = storage.Author(author)
		}

		err = <-entry.Store(roof)
		if err != nil {
			log.Print(err)
			continue
		}

		_out_entry(entry, name, err)

	}
	return true
}

func _store_dir(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// fmt.Printf("path: %s\n", path)
		if err == nil {
			if !info.IsDir() {
				parent, name := filepath.Split(path)
				if match != "" {
					if ok, _ := filepath.Match(match, name); !ok {
						log.Printf("file %s not match %s", path, match)
						return nil
					}
				}
				var tag string
				if dirAsTag {
					tag = filepath.Base(parent)
				}

				_store_file(path, roof, tag)
				return nil
			}
		} else {
			log.Printf("dir walk error: %s", err)
		}
		return err
	})
}

func _store_file(file, roof, tag string) {
	var name string
	if include_parent {
		name = _shrink_name(file)
	} else {
		name = filepath.Base(file)
	}

	// fmt.Printf("%s\n", name)
	entry, err := storage.PrepareFile(file, name)
	if err != nil {
		log.Printf("prepare file error: %s", err)
		return
	}

	qtags, err := storage.ParseTags(tag)
	if err != nil {
		log.Printf("parse tag error: %s", err)
		return
	}
	entry.Tags = qtags

	if author > 0 {
		entry.Author = storage.Author(author)
	}

	err = <-entry.Store(roof)
	if err != nil {
		log.Printf("store file error: %s", err)
	}
	_out_entry(entry, name, err)
}

func _out_entry(entry *storage.Entry, name string, err error) {
	if err != nil {
		fmt.Printf("fail %q %q\n", name, err)
	} else {
		fmt.Printf("ok %s %s %q %q\n", entry.Id, entry.Path, name, entry.Tags)
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
