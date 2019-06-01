package backend

import (
	"errors"
	"fmt"
	"github.com/go-imsto/imsto/config"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"log"
	"strings"
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

func gridfsDial(roof string) (Wagoner, error) {
	mg_url := config.GetValue(roof, "servers")
	if mg_url == "" {
		return nil, errors.New("config servers is empty")
	}
	mg_db := config.GetValue(roof, "db_name")
	if mg_db == "" {
		return nil, errors.New("config db_name is empty")
	}
	fs_prefix := config.GetValue(roof, "fs_prefix")
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
		// var err error
		g.session = g.sessionInit() //mgo.Dial(g.url)
		// if err != nil {
		// 	log.Printf("error: %s", err)
		// 	return nil, err
		// }
	}
	return g.session, nil
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
			log.Printf("OpenId(%s) error: %s", id, err)
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

func (g *gridfsConn) Put(key string, data []byte, meta JsonKV) (sev JsonKV, err error) {
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
	sev = JsonKV{"engine": "grid"}
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

var MAX_POOL_SIZE = 20

var mgoSessPool chan *mgo.Session

func (g *gridfsConn) sessionInit() *mgo.Session {
	if mgoSessPool == nil {
		mgoSessPool = make(chan *mgo.Session, MAX_POOL_SIZE)
	}
	if len(mgoSessPool) == 0 {
		go func() {
			for i := 0; i < MAX_POOL_SIZE/2; i++ {
				s, err := mgo.Dial(g.url)
				if err != nil {
					log.Printf("mgo Dial(%s) error: %s", g.url, err)
					// panic(err)
				}
				putSession(s)
			}
		}()
	}
	return <-mgoSessPool
}

func putSession(s *mgo.Session) {
	if mgoSessPool == nil {
		mgoSessPool = make(chan *mgo.Session, MAX_POOL_SIZE)
	}
	if len(mgoSessPool) >= MAX_POOL_SIZE {
		s.Close()
		return
	}
	mgoSessPool <- s
}
