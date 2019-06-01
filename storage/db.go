package storage

import (
	"database/sql"
	_ "database/sql/driver"
	_ "github.com/lib/pq"
	"log"

	"github.com/go-imsto/imsto/config"
)

var (
	dbcpool map[string]*sql.DB
)

func init() {
	dbcpool = make(map[string]*sql.DB)
}

func openDb(roof string) *sql.DB {
	dsn := config.GetValue(roof, "meta_dsn")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db error: %s", err)
	}
	return db
}

func getDb(roof string) *sql.DB {
	var db, ok = dbcpool[roof]
	if !ok {
		db = openDb(roof)
		dbcpool[roof] = db
		return db
	}

	if err := db.Ping(); err != nil {
		db = openDb(roof)
		dbcpool[roof] = db
		return db
	}

	return db
}

func withTxQuery(roof string, query func(tx *sql.Tx) error) error {

	db := getDb(roof)
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := query(tx); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
