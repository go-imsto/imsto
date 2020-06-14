package utils

import (
	"io/ioutil"
	"os"
	"path"
)

func stringInSlice(s string, a []string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}

// ReadyDir ...
func ReadyDir(filename string) error {
	dir := path.Dir(filename)
	return os.MkdirAll(dir, os.FileMode(0755))
}

// SaveFile ...
func SaveFile(filename string, data []byte) error {
	if err := ReadyDir(filename); err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, os.FileMode(0644))
}
