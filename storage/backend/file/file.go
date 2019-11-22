package backend

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/storage/backend"
)

// consts
const (
	store = "stores"
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
	repl *strings.Replacer
}

func init() {
	backend.RegisterEngine("file", locDial)
}

func locDial(roof string) (Wagoner, error) {
	dir := checkLocalDir(config.Current.LocalRoot)
	if dir == "" {
		return nil, errors.New("config local_root is empty")
	}
	logger().Debugw("locDial", "dir", dir)

	l := &locWagon{
		root: dir,
		repl: strings.NewReplacer(dir, "{root}"),
	}
	return l, nil
}

func (l *locWagon) filterError(err error) error {
	return errors.New(l.repl.Replace(err.Error()))
}

func (l *locWagon) Exists(k Key) (exist bool, err error) {
	_, err = os.Stat(path.Join(l.root, k.Path()))
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
	data, err = ioutil.ReadFile(path.Join(l.root, k.Path()))
	if err != nil {
		logger().Warnw("get fail", "key", k, "err", err)
		err = l.filterError(err)
		return
	}
	return
}

func (l *locWagon) Put(k Key, data []byte, meta Meta) (sev Meta, err error) {
	name := path.Join(l.root, k.Path())
	dir := path.Dir(name)
	err = os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		logger().Warnw("mkdirall fail", "err", err)
		err = l.filterError(err)
		return
	}
	err = ioutil.WriteFile(name, data, os.FileMode(0644))
	// sev = Meta{"root": l.root}
	if err != nil {
		logger().Warnw("write file fail", "name", name, "id", k.ID, "err", err)
		err = l.filterError(err)
		return
	}
	metaFile := name + ".meta"
	err = saveMeta(metaFile, meta)
	if err != nil {
		logger().Warnw("saveMeta fail", "metaFile", metaFile, "id", k.ID, "err", err)
		err = l.filterError(err)
		return
	}
	sev = Meta{"engine": "file", "cat": k.Cat, "size": len(data)}
	logger().Infow("save meta OK", "sev", sev, "name", name)
	return
}

func (l *locWagon) Delete(k Key) error {
	name := path.Join(l.root, k.Path())
	return os.Remove(name)
}

func checkLocalDir(dir string) string {
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger().Warnw("mkdir fail", "dir", dir, "err", err)
		return ""
	}
	return dir
}

func saveMeta(filename string, meta interface{}) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return nil
	}
	return ioutil.WriteFile(filename, data, os.FileMode(0644))
}
