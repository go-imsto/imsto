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
)

const (
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

	sql := "SELECT name, path, size, mime, meta, ids, hashes FROM " + mw.table() + " WHERE id = $1 LIMIT 1"
	entry := Entry{Id: &id}
	row := db.QueryRow(sql, id.String())

	var meta cdb.Hstore
	err := row.Scan(&entry.Name, &entry.Path, &entry.Size, &entry.Mime, &meta, &entry.Ids, &entry.Hashes)

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

	log.Printf("name: %s, path: %s, size: %d, mime: %s\n", entry.Name, entry.Path, entry.Size, entry.Mime)

	return &entry, nil
}

func (mw *MetaWrap) Store(entry *Entry) error {
	db := mw.getDb()
	defer db.Close()

	log.Printf("hashes: %s\n", entry.Hashes)
	log.Printf("ids: %s\n", entry.Ids)
	// hashes := "{" + strings.Join(entry.Hashes, ",") + "}"
	// ids := "{" + strings.Join(entry.Ids, ",") + "}"
	meta := entry.Meta.Hstore()
	log.Println(meta)
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		// tx.Rollback()
		return err
	}

	/*
	   a_section varchar,
	   	a_id varchar, a_path varchar, a_name varchar, a_mime varchar, a_size int
	   	, a_meta hstore, a_sev hstore, a_hashes varchar[], a_ids varchar[]
	*/
	// sql := "INSERT INTO " + mw.table() + "(id, name, hashes, ids, meta, path, size, mime) VALUES($1, $2, $3, $4, $5, $6, $7, $8)"
	// result, err := tx.Exec(sql, entry.Id.String(), entry.Name, entry.Hashes, entry.Ids, meta, entry.Path, entry.Size, entry.Mime)
	sql := "SELECT entry_save($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);"
	row := tx.QueryRow(sql, mw.table_suffix,
		entry.Id.String(), entry.Path, entry.Name, entry.Mime, entry.Size, meta, entry.sev, entry.Hashes, entry.Ids)

	var ret int
	err = row.Scan(&ret)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("entry save ret: %v\n", ret)

	// var ra int64
	// ra, err = result.RowsAffected()

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println("RowsAffected ", ra)

	tx.Commit()

	return err
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
