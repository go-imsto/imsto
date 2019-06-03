package storage

import (
	"database/sql"
	_ "github.com/lib/pq"
	"os"

	zlog "github.com/go-imsto/imsto/log"
)

var (
	dbcpool map[string]*sql.DB

	dbDSN string
)

func logger() zlog.Logger {
	return zlog.Get()
}

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
