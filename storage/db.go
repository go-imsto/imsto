package storage

import (
	"database/sql"
	"os"
	"sync"

	_ "github.com/lib/pq" // just ok
)

type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type DBer interface {
	Queryer
	Begin() (*sql.Tx, error)
}

type DBTxer interface {
	Queryer
	Rollback() error
	Commit() error
}

var (
	dbcpool map[string]*sql.DB
	dbmu    sync.Mutex

	dbDSN string
)

func init() {
	if s, exists := os.LookupEnv("IMSTO_META_DSN"); exists && s != "" {
		dbDSN = s
	} else {
		dbDSN = "postgres://imsto@localhost/imsto?sslmode=disable"
	}

	dbcpool = make(map[string]*sql.DB)

}

func openDb(roof string) *sql.DB {
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		logger().Fatalw("open db fail", "err", err)
	}
	return db
}

func getDb(roof string) *sql.DB {
	dbmu.Lock()
	defer dbmu.Unlock()
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

func envOr(key, dft string) string {
	v := os.Getenv(key)
	if v == "" {
		return dft
	}
	return v
}
