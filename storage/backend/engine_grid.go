package backend

import (
	"errors"
	"fmt"
	"io/ioutil"
	"labix.org/v2/mgo"
	"log"
	"strings"
	"wpst.me/calf/config"
	"wpst.me/calf/db"
)

type gridfsConn struct {
	url, db, prefix string
	fs              *mgo.GridFS
	session         *mgo.Session
}

func init() {
	RegisterEngine("grid", gridfsDial)
	RegisterEngine("mongodb", gridfsDial) // for old imsto config
}

func gridfsDial(sn string) (Wagoner, error) {
	mg_url := config.GetValue(sn, "servers")
	if mg_url == "" {
		return nil, errors.New("config servers is empty")
	}
	mg_db := config.GetValue(sn, "db_name")
	if mg_db == "" {
		return nil, errors.New("config db_name is empty")
	}
	fs_prefix := config.GetValue(sn, "fs_prefix")
	if fs_prefix == "" {
		return nil, errors.New("config fs_prefix is empty")
	}
	g := &gridfsConn{
		url:    mg_url,
		db:     mg_db,
		prefix: fs_prefix,
	}
	return g, nil
}

func (g *gridfsConn) getSession() (*mgo.Session, error) {
	if g.session == nil {
		var err error
		g.session, err = mgo.Dial(g.url)
		if err != nil {
			log.Printf("error: %s", err)
			return nil, err
		}
	}
	return g.session.Clone(), nil
}

func (g *gridfsConn) withFs(f func(*mgo.GridFS) error) error {
	session, err := g.getSession()
	if err != nil {
		return err
	}
	defer session.Close()
	fs := session.DB(g.db).GridFS(g.prefix)
	return f(fs)
}

func (g *gridfsConn) Exists(key string) (exist bool, err error) {
	id := pathToId(key)
	c := func(fs *mgo.GridFS) error {
		r, err := fs.OpenId(id)
		if err != nil {
			return err
		}
		defer r.Close()
		exist = true
		return nil
	}
	err = g.withFs(c)
	return
}

func (g *gridfsConn) Get(key string) (data []byte, err error) {
	id := pathToId(key)
	c := func(fs *mgo.GridFS) error {
		r, err := fs.OpenId(id)
		if err != nil {
			return err
		}
		defer r.Close()
		data, err = ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		return nil
	}
	err = g.withFs(c)
	return
}

func (g *gridfsConn) Put(key string, data []byte, meta db.Hstore) (sev db.Hstore, err error) {
	id := pathToId(key)
	c := func(fs *mgo.GridFS) error {
		f, err := fs.Create(key)
		if err != nil {
			return err
		}
		defer f.Close()
		f.SetId(id)
		f.SetMeta(meta)
		f.SetContentType(fmt.Sprint(meta.Get("mime")))
		var n int
		n, err = f.Write(data)
		if err != nil {
			return err
		}
		log.Printf("gridfs wrote %d", n)
		return nil
	}
	err = g.withFs(c)
	sev = db.Hstore{"engine": "grid"}
	return
}

func (g *gridfsConn) Del(key string) error {
	id := pathToId(key)
	c := func(fs *mgo.GridFS) error {
		return fs.Remove(id)
	}
	return g.withFs(c)
}

func pathToId(key string) string {
	r := strings.NewReplacer("/", "")
	key = r.Replace(key)
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == '.' {
			return key[0:i]
		}
	}
	return key
}
