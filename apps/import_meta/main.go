package main

import (
	"flag"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"time"
	"wpst.me/calf/db"
	"wpst.me/calf/storage"
)

// url: mongodb://db.wp.net,db20.wp.net/storage
// url: localhost

var (
	mgo_url     string
	mgo_db      string
	roof        string
	mgo_coll    string
	skip, limit int
	mgoSession  *mgo.Session
)

type entryOut struct {
	Id      string    `bson:"_id,omitempty" json:"id"`
	Name    string    `bson:"name"`
	Path    string    `bson:"filename"`
	Mime    string    `bson:"mime,omitempty"`
	Size    uint32    `bson:"size"`
	Hashes  []string  `bson:"hash"`
	Ids     []string  `bson:"ids"`
	Meta    db.Hstore `bson:"meta",omitempty`
	Sev     db.Hstore `bson:"sev",omitempty`
	Created time.Time `bson:"created,omitempty"`
}

func (eo entryOut) toEntry() (entry *storage.Entry, err error) {
	eo.Meta.Set("mime", eo.Mime)
	entry, err = storage.NewEntryConvert(eo.Id, eo.Name, eo.Path, eo.Size, eo.Meta, eo.Sev, eo.Hashes, eo.Ids, eo.Created)
	if err != nil {
		log.Printf("toEntry error: %s", err)
		return
	}
	return
}

func (eo entryOut) save() error {
	entry, err := eo.toEntry()
	log.Printf("import %s %s %d", entry.Id, entry.Path, entry.Size)
	return err
}

func init() {
	flag.StringVar(&mgo_url, "url", "mongodb://localhost/storage", "mongodb server url")
	flag.StringVar(&mgo_db, "db", "storage", "mongodb database name")
	flag.StringVar(&roof, "roof", "s3", "mongodb collection name")
	flag.IntVar(&skip, "skip", 0, "skip")
	flag.IntVar(&limit, "limit", 5, "limit")
	flag.Parse()
	mgo_coll = roof + ".files"
}

func main() {
	q := bson.M{}
	total, err := CountEntry(mgo_coll, q)
	if err != nil {
		log.Printf("count error: %s", err)
		return
	}
	log.Printf("total: %d", total)
	// skip := 0
	// limit := 5
	for skip < total {
		results, err := QueryEntry(mgo_coll, q, skip, limit)
		if err != nil {
			log.Printf("query error: %s", err)
		}
		// log.Printf("results: %s", results)
		for i, e := range results {
			// log.Printf("%d %s\n", i, e.toEntry())
			err = e.save()
			if err != nil {
				log.Printf("save %d error: %s", i, err)
			}
		}
		skip += limit
	}
}

func getSession() (*mgo.Session, error) {
	if mgoSession == nil {
		var err error
		mgoSession, err = mgo.Dial(mgo_url)
		if err != nil {
			log.Printf("error: %s", err)
			return nil, err
		}
	}
	return mgoSession.Clone(), nil
}

func withCollection(collection string, s func(*mgo.Collection) error) error {
	session, err := getSession()
	if err != nil {
		return err
	}
	defer session.Close()
	c := session.DB(mgo_db).C(collection)
	// log.Printf("connection: %v", c)
	return s(c)
}

func QueryEntry(collection string, q interface{}, skip int, limit int) (results []entryOut, err error) {
	results = []entryOut{}
	query := func(c *mgo.Collection) error {
		fn := c.Find(q).Skip(skip).Limit(limit).All(&results)
		if limit < 0 {
			fn = c.Find(q).Skip(skip).All(&results)
		}
		return fn
	}
	search := func() error {
		return withCollection(collection, query)
	}
	err = search()
	return
}

func CountEntry(collection string, q interface{}) (n int, err error) {
	query := func(c *mgo.Collection) error {
		n, err = c.Count()
		return err
	}
	count := func() error {
		return withCollection(collection, query)
	}
	err = count()
	return
}
