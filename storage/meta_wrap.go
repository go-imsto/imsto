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
	Browse(limit int, offset int) (*sql.Rows, error)
	Store(entry *Entry) error
	Get(id EntryId) (*Entry, error)
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

func NewMetaWrapper(section string) (mw MetaWrapper) {
	mw = newMetaWrap(section)
	return mw
}

func (mw *MetaWrap) table() string {
	return meta_table_prefix + mw.table_suffix
}

func (mw *MetaWrap) Browse(limit int, offset int) (rows *sql.Rows, err error) {
	if limit < 1 {
		limit = 1
	}
	if offset < 0 {
		offset = 0
	}

	db := mw.getDb()

	rows, err = db.Query("SELECT * FROM "+mw.table()+" LIMIT $1 OFFSET $2", limit, offset)

	log.Print(rows)
	return rows, err
}

func (mw *MetaWrap) Get(id EntryId) (*Entry, error) {
	db := mw.getDb()
	defer db.Close()

	sql := "SELECT name, path, size, meta, sev, ids, hashes FROM " + mw.table() + " WHERE id = $1 LIMIT 1"
	entry := Entry{Id: &id}
	row := db.QueryRow(sql, id.String())

	var meta, sev cdb.Hstore
	err := row.Scan(&entry.Name, &entry.Path, &entry.Size, &meta, &sev, &entry.Ids, &entry.Hashes)

	if err != nil {
		log.Println(err)
		return &entry, err
	}

	log.Println("first id:", entry.Ids[0])
	// entry.Meta = image.
	log.Println("meta:", meta)
	var ia image.ImageAttr
	err = meta.ToStruct(&ia)
	if err != nil {
		log.Println(err)
		return &entry, err
	}
	log.Println(ia.Width)

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

func (mw *MetaWrap) ExistHash(hash string) string {
	db := mw.getDb()
	defer db.Close()
	var id string
	sql := "SELECT item_id FROM " + tableHash(hash) + " WHERE hashed = $1 LIMIT 1"
	row := db.QueryRow(sql, hash)
	err := row.Scan(&id)
	if err != nil {
		log.Println(err)
		return ""
	}
	return id
}

func (mw *MetaWrap) ExistMap(id string) string {
	db := mw.getDb()
	defer db.Close()
	var eid string
	sql := "SELECT id FROM " + tableMap(id) + " WHERE id = $1 LIMIT 1"
	row := db.QueryRow(sql, id)
	err := row.Scan(&eid)
	if err != nil {
		log.Println(err)
		return ""
	}
	return eid
}

func (mw *MetaWrap) Delete(id EntryId) error {
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
