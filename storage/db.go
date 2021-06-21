package storage

import (
	"database/sql"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq" // just ok
)

// Queryer ...
type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// DBer ...
type DBer interface {
	Queryer
	Begin() (*sql.Tx, error)
}

// DBTxer ...
type DBTxer interface {
	Queryer
	Rollback() error
	Commit() error
}

var (
	dbc          *sql.DB
	once         sync.Once
	quitC        chan struct{}
	pingInterval = 90 * time.Second

	dbDSN string
)

func init() {
	dbDSN = envOr("IMSTO_META_DSN", "postgres://imsto@localhost/imsto?sslmode=disable")

	quitC = make(chan struct{})
}

func openDb() *sql.DB {
	logger().Infow("openDb")
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		logger().Fatalw("open db fail", "err", err)
	}
	return db
}

// Close database close, stop ping
func Close() {
	close(quitC)
	if dbc != nil {
		logger().Infow("closeDb")
		dbc.Close()
	}
}

func getDb() *sql.DB {
	if dbc == nil {
		once.Do(func() {
			dbc = openDb()
			go reap(pingInterval, pingDb, quitC)
		})
	}
	return dbc
}

func pingDb() error {
	err := dbc.Ping()
	if err != nil {
		logger().Infow("ping db fail, reconnect", "err", err)
		dbc = openDb()
	}
	return err
}

// reap with special action at set intervals.
func reap(interval time.Duration, cf func() error, quit <-chan struct{}) {
	logger().Debugw("starting reaper", "interval", interval)
	ticker := time.NewTicker(interval)

	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-quit:
			// Handle the quit signal.
			return
		case <-ticker.C:
			// Execute function of clean.
			if err := cf(); err != nil {
				logger().Infow("reap fail", "err", err)
			}
		}
	}
}

func withTxQuery(query func(tx *sql.Tx) error) error {

	db := getDb()

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
