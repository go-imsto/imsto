package storage

import (
	"calf/config"
	// "calf/image"
	"database/sql"
	_ "database/sql/driver"
	_ "github.com/lib/pq"
	"log"
	"strings"
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

	table string
}

func newMetaWrap(section string) *MetaWrap {
	if section == "" {
		section = "common"
	}
	dsn := config.GetValue(section, "meta_dsn")
	table := config.GetValue(section, "meta_table")
	mw := &MetaWrap{dsn: dsn, table: table}

	return mw
}

func NewMetaWrapper(section string) (mw MetaWrapper) {
	mw = newMetaWrap(section)
	return mw
}

func (mw *MetaWrap) Browse(limit int, offset int) (rows *sql.Rows, err error) {
	if limit < 1 {
		limit = 1
	}
	if offset < 0 {
		offset = 0
	}

	db := mw.getDb()

	table := mw.table

	rows, err = db.Query("SELECT * FROM "+table+" LIMIT $1 OFFSET $2", limit, offset)

	log.Print(rows)
	return rows, err
}

func (mw *MetaWrap) Get(id EntryId) (*Entry, error) {
	db := mw.getDb()
	defer db.Close()

	table := mw.table
	sql := "SELECT name, path, size, mime, meta FROM " + table + " WHERE id = $1 LIMIT 1"
	entry := Entry{Id: &id}
	row := db.QueryRow(sql, id.String())

	var meta Hstore
	err := row.Scan(&entry.Name, &entry.Path, &entry.Size, &entry.Mime, &meta)

	if err != nil {
		log.Println(err)
		return &entry, err
	}
	// entry.Meta = image.
	log.Println("meta:", meta)

	log.Printf("name: %s, path: %s, size: %d, mime: %s\n", entry.Name, entry.Path, entry.Size, entry.Mime)

	return &entry, nil
}

func (mw *MetaWrap) Store(entry *Entry) error {
	db := mw.getDb()
	defer db.Close()

	table := mw.table
	log.Println("table", table)
	hashes := "{" + strings.Join(entry.Hashes, ",") + "}"
	ids := "{" + strings.Join(entry.Ids, ",") + "}"
	meta := ia2hstore(entry.Meta).String()
	log.Println(meta)
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		// tx.Rollback()
		return err
	}

	sql := "INSERT INTO " + table + "(id, name, hashes, ids, meta, path, size, mime) VALUES($1, $2, $3, $4, $5, $6, $7, $8)"
	result, err := tx.Exec(sql, entry.Id.String(), entry.Name, hashes, ids, meta, entry.Path, entry.Size, entry.Mime)

	if err != nil {
		log.Fatal(err)
	}

	var ra int64
	ra, err = result.RowsAffected()

	if err != nil {
		log.Fatal(err)
	}

	log.Println("RowsAffected ", ra)

	tx.Commit()

	return nil
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
