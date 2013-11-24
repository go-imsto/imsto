package storage

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"wpst.me/calf/config"
	"wpst.me/calf/db"
)

// local storage wagon
type locWagon struct {
	root string
}

func init() {
	RegisterEngine("file", locDial)
}

func locDial(sn string) (Wagoner, error) {
	dir := config.GetValue(sn, "local_root")
	if dir == "" {
		return nil, errors.New("config local_root is empty")
	}
	l := &locWagon{
		root: dir,
	}
	return l, nil
}

func (l *locWagon) Exists(key string) (exist bool, err error) {
	// var fi os.FileInfo
	name := path.Join(l.root, key)
	_, err = os.Stat(name)
	if os.IsNotExist(err) {
		exist = false
	}
	exist = true
	return
}

func (l *locWagon) Get(key string) (data []byte, err error) {
	name := path.Join(l.root, key)
	data, err = ioutil.ReadFile(name)
	return
}

func (l *locWagon) Put(key string, data []byte, meta db.Hstore) (sev db.Hstore, err error) {
	name := path.Join(l.root, key)
	dir := path.Dir(name)
	err = os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		return
	}
	err = ioutil.WriteFile(name, data, os.FileMode(0644))
	// sev = db.Hstore{"root": l.root}

	return
}

func (l *locWagon) Del(key string) error {
	name := path.Join(l.root, key)
	return os.Remove(name)
}
