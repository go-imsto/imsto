package imsto

import (
	"database/sql"
	_ "database/sql/driver"
	_ "github.com/lib/pq"
	"log"
)

type MetaWrapper interface {
	Browse(limit int, offset int) (*sql.Rows, error)
	Store(entry *Entry) error
	Delete(id EntryId) error
}

type MetaWrap struct {
	dsn string
	// section string
	// prefix  string

	table string
}

func NewMetaWrapper(section string) (*MetaWrap, error) {
	if section == "" {
		section = "common"
	}
	dsn := getConfig(section, "meta_dsn")
	table := getConfig(section, "meta_table")
	mw := MetaWrap{dsn: dsn, table: table}

	return &mw, nil
}

func (mw *MetaWrap) Browse(limit int, offset int) (rows *sql.Rows, err error) {
	if limit < 1 {
		limit = 1
	}
	if offset < 0 {
		offset = 0
	}

	db := mw.getDb()

	rows, err = db.Query("SELECT * FROM "+mw.table+" LIMIT $1 OFFSET $2", limit, offset)

	log.Print(rows)
	return rows, err
}

func (mw *MetaWrap) Store(entry *Entry) error {
	// db := mw.getDb()
	// db.Begin()
	// db.Exec("INSERT INTO "+mw.table+"(id, name, hashes, ids, path, mime size)", ...)

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
