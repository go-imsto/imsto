package backend

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"

	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/storage/backend"
)

// consts
const (
	name  = "stores"
	thumb = "thumb"
)

// Wagoner ...
type Wagoner = backend.Wagoner

// ListSpec ...
type ListSpec = backend.ListSpec

// ListItem ...
type ListItem = backend.ListItem

// Key ...
type Key = backend.Key

// Meta ...
type Meta = backend.Meta

// local storage wagon
type locWagon struct {
	root string
}

func init() {
	backend.RegisterEngine("file", locDial)
}

func locDial(roof string) (Wagoner, error) {
	dir := checkLocalDir(config.Current.LocalRoot)
	if dir == "" {
		return nil, errors.New("config local_root is empty")
	}

	l := &locWagon{
		root: dir,
	}
	return l, nil
}

func (l *locWagon) Exists(k Key) (exist bool, err error) {
	name := path.Join(l.root, k.Path())
	_, err = os.Stat(name)
	if os.IsNotExist(err) {
		exist = false
	}
	exist = true
	return
}

func (l *locWagon) List(ls ListSpec) (items []ListItem, err error) {
	// TODO:
	return
}

func (l *locWagon) Get(k Key) (data []byte, err error) {
	name := path.Join(l.root, k.Path())
	data, err = ioutil.ReadFile(name)
	return
}

func (l *locWagon) Put(k Key, data []byte, meta Meta) (sev Meta, err error) {
	name := path.Join(l.root, k.Path())
	dir := path.Dir(name)
	err = os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		return
	}
	err = ioutil.WriteFile(name, data, os.FileMode(0644))
	// sev = Meta{"root": l.root}
	if err != nil {
		logger().Warnw("write file fail", "name", name, "id", k.ID, "err", err)
		return
	}
	metaFile := name + ".meta"
	err = saveMeta(metaFile, meta)
	if err != nil {
		logger().Warnw("saveMeta fail", "metaFile", metaFile, "id", k.ID, "err", err)
		return
	}
	sev = Meta{"engine": "file", "key": k, "size": len(data)}
	logger().Infow("save meta OK", "sev", sev, "name", name)
	return
}

func (l *locWagon) Delete(k Key) error {
	name := path.Join(l.root, k.Path())
	return os.Remove(name)
}

func checkLocalDir(dir string) string {
	if err := os.Mkdir(dir, 0755); err == nil {
		return dir
	}
	logger().Warnw("mkdir fail", "dir", dir)
	return ""
}

func saveMeta(filename string, meta interface{}) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return nil
	}
	return ioutil.WriteFile(filename, data, os.FileMode(0644))
}
