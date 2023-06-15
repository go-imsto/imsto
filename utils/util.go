package utils

import (
	"os"
	"path"
)

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
	return os.WriteFile(filename, data, os.FileMode(0644))
}

// Exists returns true if a file exists
func Exists(fpath string) bool {
	_, err := os.Stat(fpath)
	return !os.IsNotExist(err)
}

// IsDir ...
func IsDir(fpath string) bool {
	fi, err := os.Stat(fpath)
	return err == nil && fi.Mode().IsDir()
}

// IsRegular ...
func IsRegular(fpath string) bool {
	fi, err := os.Stat(fpath)
	return err == nil && fi.Mode().IsRegular()
}

// FileSize return file size, return -1 if error
func FileSize(fpath string) int64 {
	if fi, err := os.Stat(fpath); err == nil {
		return fi.Size()
	}
	return -1
}
