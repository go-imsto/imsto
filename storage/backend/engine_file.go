package backend

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/go-imsto/imsto/config"
)

// local storage wagon
type locWagon struct {
	root string
}

func init() {
	RegisterEngine("file", locDial)
}

func locDial(roof string) (Wagoner, error) {
	dir := config.GetValue(roof, "local_root")
	if dir == "" {
		dir = _check_local_dir()
		if dir == "" {
			return nil, errors.New("config local_root is empty")
		}

	}
	l := &locWagon{
		root: dir,
	}
	return l, nil
}

func (l *locWagon) Exists(id string) (exist bool, err error) {
	name := path.Join(l.root, Id2Path(id))
	_, err = os.Stat(name)
	if os.IsNotExist(err) {
		exist = false
	}
	exist = true
	return
}

func (l *locWagon) Get(id string) (data []byte, err error) {
	name := path.Join(l.root, Id2Path(id))
	data, err = ioutil.ReadFile(name)
	return
}

func (l *locWagon) Put(id string, data []byte, meta JsonKV) (sev JsonKV, err error) {
	key := Id2Path(id)
	name := path.Join(l.root, key)
	dir := path.Dir(name)
	err = os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		return
	}
	err = ioutil.WriteFile(name, data, os.FileMode(0644))
	// sev = JsonKV{"root": l.root}
	if err != nil {
		return
	}
	meta_file := name + ".meta"
	err = saveMeta(meta_file, meta)
	if err != nil {
		log.Print("engine file save meta fail: ", err.Error())
	}
	sev = JsonKV{"engine": "file", "key": key}
	log.Printf("engine file save %s done", key)
	return
}

func (l *locWagon) Del(id string) error {
	name := path.Join(l.root, Id2Path(id))
	return os.Remove(name)
}

func _exists_dir(dir string) bool {
	if fi, err := os.Stat(dir); err == nil {
		if fi.IsDir() {
			return true
		}
	}
	return false
}

func _check_local_dir() string {
	if _home := os.Getenv("HOME"); _home != "" {
		// check darwin User Library
		_dir := path.Join(_home, "Library")
		if _exists_dir(_dir) {
			_dir = path.Join(_home, "Libarry", "imsto")
			if _exists_dir(_dir) {
				return _dir
			} else {
				if err := os.Mkdir(_dir, 0755); err == nil {
					return _dir
				}
			}
		}
	}
	return ""
}

func saveMeta(filename string, meta interface{}) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return nil
	}
	return ioutil.WriteFile(filename, data, os.FileMode(0644))
}
