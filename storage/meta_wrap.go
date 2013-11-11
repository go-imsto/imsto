package storage

import (
	"calf/config"
	cdb "calf/db"
	"calf/image"
	"database/sql"
	_ "database/sql/driver"
	_ "github.com/lib/pq"
	"log"
	// "strings"
	"fmt"
)

const (
	hash_table_prefix = "hash_"
	map_table_prefix  = "mapping_"
	meta_table_prefix = "meta_"
)

type MetaWrapper interface {
	Browse(limit, offset int) ([]Entry, int, error)
	Store(entry *Entry) error
	GetMeta(id EntryId) (*Entry, error)
	GetHash(hash string) (*ehash, error)
	GetMap(id EntryId) (*emap, error)
	Delete(id EntryId) error
}

type MetaWrap struct {
	dsn string
	// section string
	// prefix  string

	table_suffix string
}

func newMetaWrap(section string) *MetaWrap {
	if section == "" {
		section = "common"
	}
	dsn := config.GetValue(section, "meta_dsn")
	table := config.GetValue(section, "meta_table_suffix")
	mw := &MetaWrap{dsn: dsn, table_suffix: table}

	return mw
}

var meta_wrappers = make(map[string]MetaWrapper)

func NewMetaWrapper(section string) (mw MetaWrapper) {
	var ok bool
	if mw, ok = meta_wrappers[section]; !ok {
		mw = newMetaWrap(section)
		meta_wrappers[section] = mw
	}

	return mw
}

func (mw *MetaWrap) table() string {
	return meta_table_prefix + mw.table_suffix
}

func (mw *MetaWrap) Browse(limit, offset int) (a []Entry, t int, err error) {
	if limit < 1 {
		limit = 1
	}
	if offset < 0 {
		offset = 0
	}

	db := mw.getDb()
	defer db.Close()

	t = 0
	err = db.QueryRow("SELECT COUNT(id) FROM " + mw.table() + "").Scan(&t)
	if err != nil {
		return
	}

	if t == 0 {
		log.Print("empty meta")
		return
	}

	var r *sql.Rows
	r, err = db.Query("SELECT id, name, path, size, meta, sev, ids, hashes FROM "+mw.table()+" LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return
	}

	defer r.Close()

	for r.Next() {
		e := Entry{}
		var id string
		var meta, sev cdb.Hstore
		err = r.Scan(&id, &e.Name, &e.Path, &e.Size, &meta, &sev, &e.Ids, &e.Hashes)
		if err != nil {
			return
		}
		e.Id, err = NewEntryId(id)
		if err != nil {
			return
		}

		var ia image.ImageAttr
		err = meta.ToStruct(&ia)
		if err != nil {
			return
		}

		e.Meta = &ia
		e.Mime = fmt.Sprint(meta.Get("mime"))

		a = append(a, e)
	}
	return
}

func (mw *MetaWrap) GetMeta(id EntryId) (*Entry, error) {
	db := mw.getDb()
	defer db.Close()

	sql := "SELECT name, path, size, meta, sev, ids, hashes FROM " + mw.table() + " WHERE id = $1 LIMIT 1"
	entry := Entry{Id: &id}
	row := db.QueryRow(sql, id.String())

	var meta cdb.Hstore
	err := row.Scan(&entry.Name, &entry.Path, &entry.Size, &meta, &entry.sev, &entry.Ids, &entry.Hashes)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println("first id:", entry.Ids[0])

	log.Println("meta:", meta)
	log.Println("sev:", entry.sev)
	var ia image.ImageAttr
	err = meta.ToStruct(&ia)
	if err != nil {
		log.Println(err)
		return &entry, err
	}
	log.Println("ia:", ia)

	entry.Meta = &ia
	entry.Mime = fmt.Sprint(meta.Get("mime"))

	log.Printf("name: %s, path: %s, size: %d, mime: %s\n", entry.Name, entry.Path, entry.Size, entry.Mime)

	return &entry, nil
}

func (mw *MetaWrap) Store(entry *Entry) error {
	db := mw.getDb()
	defer db.Close()

	log.Printf("hashes: %s\n", entry.Hashes)
	log.Printf("ids: %s\n", entry.Ids)

	meta := entry.Meta.Hstore()
	log.Println("meta: ", meta)
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		// tx.Rollback()
		return err
	}

	sql := "SELECT entry_save($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);"
	row := tx.QueryRow(sql, mw.table_suffix,
		entry.Id.String(), entry.Path, entry.Name, entry.Mime, entry.Size, meta, entry.sev, entry.Hashes, entry.Ids)

	var ret int
	err = row.Scan(&ret)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("entry save ret: %v\n", ret)

	tx.Commit()

	return err
}

type ehash struct {
	hash, path, id string
}

func (mw *MetaWrap) GetHash(hash string) (*ehash, error) {
	db := mw.getDb()
	defer db.Close()
	var e ehash
	sql := "SELECT item_id, path FROM " + tableHash(hash) + " WHERE hashed = $1 LIMIT 1"
	row := db.QueryRow(sql, hash)
	err := row.Scan(&e.id, &e.path)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	e.hash = hash
	return &e, nil
}

type emap struct {
	id               EntryId
	name, path, mime string
	size             uint32
	sev              cdb.Hstore
	status           uint16
}

func (mw *MetaWrap) GetMap(id EntryId) (*emap, error) {
	db := mw.getDb()
	defer db.Close()
	sql := "SELECT name, path, mime, size, sev, status FROM " + tableMap(id.String()) + " WHERE id = $1 LIMIT 1"
	row := db.QueryRow(sql, id.String())
	var e = emap{id: id}
	err := row.Scan(&e.name, &e.path, &e.mime, &e.size, &e.sev, &e.status)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &e, nil
}

func (mw *MetaWrap) Delete(id EntryId) error {
	// TODO:
	return nil
}

func (mw *MetaWrap) getDb() *sql.DB {
	dsn := mw.dsn
	db, err := sql.Open("postgres", dsn)

	if err != nil {
		log.Fatal(err)
	}
	return db
}

func tableHash(s string) string {
	return hash_table_prefix + s[:1]
}

func tableMap(s string) string {
	return map_table_prefix + s[:1]
}
