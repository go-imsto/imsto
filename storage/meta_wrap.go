package storage

import (
	"database/sql"
	_ "database/sql/driver"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/image"
	cdb "github.com/go-imsto/imsto/storage/types"
)

const (
	hash_table_prefix = "hash_"
	map_table_prefix  = "mapping_"
	meta_table_prefix = "meta_"
	max_args          = 10
)

type MetaWrapper interface {
	Browse(limit, offset int, sort map[string]int, filter MetaFilter) ([]*Entry, error)
	Count(filter MetaFilter) (int, error)
	Ready(entry *Entry) error
	SetDone(id EntryId, sev cdb.JsonKV) error
	Save(entry *Entry, is_update bool) error
	BatchSave(entries []*Entry) error
	GetMeta(id EntryId) (*Entry, error)
	GetHash(hash string) (*ehash, error)
	GetEntry(id EntryId) (*Entry, error)
	Delete(id EntryId) error
	MapTags(id EntryId, tags string) error
	UnmapTags(id EntryId, tags string) error
}

type rowScanner interface {
	Scan(...interface{}) error
}

type MetaWrap struct {
	roof string
	// prefix  string

	table_suffix string
}

func newMetaWrap(roof string) *MetaWrap {
	if roof == "" {
		log.Print("arg roof is empty")
		roof = "common"
	}

	table := config.GetValue(roof, "meta_table_suffix")
	if table == "" {
		// log.Print("meta_table_suffix is empty, use roof")
		table = roof
	}
	log.Printf("table suffix: %s", table)
	mw := &MetaWrap{roof: roof, table_suffix: table}

	return mw
}

var (
	meta_wrappers   = make(map[string]MetaWrapper)
	meta_columns    = "id, path, name, meta, hashes, ids, size, sev, tags, exif, app_id, author, created, roof"
	sortable_fields = []string{"id", "created"}
	ErrDbError      = errors.New("database error")
)

func NewMetaWrapper(roof string) (mw MetaWrapper) {
	var ok bool
	if mw, ok = meta_wrappers[roof]; !ok {
		mw = newMetaWrap(roof)
		meta_wrappers[roof] = mw
	}

	return mw
}

func (mw *MetaWrap) TableSuffix() string {
	return mw.table_suffix
}

func (mw *MetaWrap) table() string {
	return meta_table_prefix + mw.table_suffix
}

const (
	ASCENDING  = 1
	DESCENDING = -1
)

func isSortable(k string) bool {
	for _, sf := range sortable_fields {
		if k == sf {
			return true
		}
	}
	return false
}

type MetaFilter struct {
	Tags   string
	App    AppId
	Author Author
}

func buildWhere(filter MetaFilter) (where string, args []interface{}) {
	where = " WHERE status = 0"

	argc := 0
	args = make([]interface{}, 0, max_args)

	var qtags, _ = cdb.NewQarrayText(filter.Tags)

	if len(qtags) > 0 {
		log.Printf("qtags len: %d", len(qtags))
		where = fmt.Sprintf("%s AND tags @> $%d", where, argc+1)
		args = append(args, qtags)
		argc++
	}

	var author = filter.Author
	if author > 0 {
		log.Printf("author: %d", author)
		where = fmt.Sprintf("%s AND author = $%d", where, argc+1)
		args = append(args, author)
		argc++
	}

	var app = filter.App
	if app > 0 {
		log.Printf("app: %d", app)
		where = fmt.Sprintf("%s AND app_id = $%d", where, argc+1)
		args = append(args, app)
		argc++
	}

	return
}

func (mw *MetaWrap) Count(filter MetaFilter) (t int, err error) {
	db := mw.getDb()
	table := mw.table()

	t = 0
	where, args := buildWhere(filter)
	// err = db.QueryRow("SELECT COUNT(id) FROM "+table+where, args...).Scan(&t)
	rows, err := db.Query("SELECT COUNT(id) FROM "+table+where, args...)
	// return &Row{rows: rows, err: err}
	if err != nil {
		logger().Warnw("query count fail", "err", err)
		err = ErrDbError
		return
	}

	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&t)
	}

	err = rows.Err()

	return
}

func (mw *MetaWrap) Browse(limit, offset int, sort map[string]int, filter MetaFilter) (a []*Entry, err error) {
	if limit < 1 {
		limit = 1
	}
	if offset < 0 {
		offset = 0
	}
	table := mw.table()

	log.Printf("browse table: %s", table)

	db := mw.getDb()

	var orders []string
	for k, v := range sort {
		if isSortable(k) {
			var o string
			if v == ASCENDING {
				o = "ASC"
			} else {
				o = "DESC"
			}
			orders = append(orders, k+" "+o)
		}
	}
	where, args := buildWhere(filter)
	str := "SELECT " + meta_columns + " FROM " + table + where

	if len(orders) > 0 {
		str = str + " ORDER BY " + strings.Join(orders, ",")
	}
	// log.Printf("sql: %s", str)
	var r *sql.Rows
	argc := len(args)
	r, err = db.Query(fmt.Sprintf("%s LIMIT $%d OFFSET $%d", str, argc+1, argc+2), append(args, limit, offset)...)
	if err != nil {
		logger().Warnw("browse fail", "err", err)
		err = ErrDbError
		return
	}

	defer r.Close()

	for r.Next() {
		var entry *Entry
		entry, err = _bindRow(r)
		if err != nil {
			return
		}
		a = append(a, entry)
	}
	return
}

// depends meta_columns
func _bindRow(rs rowScanner) (*Entry, error) {

	e := Entry{}
	var id, roof string
	var meta image.Attr
	// "id, path, name, meta, hashes, ids, size, sev, exif, app_id, author, created, roof"
	err := rs.Scan(&id, &e.Path, &e.Name, &meta, &e.Hashes, &e.Ids, &e.Size,
		&e.sev, &e.Tags, &e.exif, &e.AppId, &e.Author, &e.Created, &roof)
	if err != nil {
		logger().Warnw("bind fail", "err", err)
		err = ErrDbError
		return nil, err
	}
	e.Id, err = NewEntryId(id)
	if err != nil {
		return nil, err
	}

	// log.Printf("bind meta %s: %v", meta, ia)

	e.Meta = &meta
	e.Mime = meta.Mime
	e.Roofs = cdb.StringArray{roof}

	// log.Printf("bind name: %s, path: %s, size: %d, mime: %s\n", e.Name, e.Path, e.Size, e.Mime)

	return &e, nil
}

func (mw *MetaWrap) GetMeta(id EntryId) (*Entry, error) {
	db := mw.getDb()

	sql := "SELECT " + meta_columns + " FROM " + mw.table() + " WHERE id = $1 LIMIT 1"

	row := db.QueryRow(sql, id.String())

	return _bindRow(row)
}

func popPrepared() (*Entry, error) {
	mwr := NewMetaWrapper("common")
	mw := mwr.(*MetaWrap)

	db := mw.getDb()

	sql := "SELECT " + meta_columns + " FROM meta__prepared ORDER BY created ASC LIMIT 1"

	row := db.QueryRow(sql)

	return _bindRow(row)
}

func (mw *MetaWrap) Ready(entry *Entry) error {
	return mw.withTxQuery(func(tx *sql.Tx) error {
		var created time.Time
		err := tx.QueryRow("SELECT created FROM meta__prepared WHERE id = $1", entry.Id).Scan(&created)
		if err == nil {
			return nil
		}
		logger().Infow("load prepared fail", "id", entry.Id, "err", err)

		_, err = tx.Exec(`INSERT INTO meta__prepared (id, roof, path, name, size, meta, hashes, ids, app_id, author, tags)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`, entry.Id, mw.table_suffix, entry.Path,
			entry.Meta.Name, entry.Meta.Size, entry.Meta, entry.Hashes, entry.Ids,
			entry.AppId, entry.Author, entry.Tags)
		if err != nil {
			logger().Warnw("save prepared fail", "id", entry.Id, "err", err)
		} else {
			logger().Infow("save prepared OK", "id", entry.Id)
		}
		return err
	})
}

func (mw *MetaWrap) SetDone(id EntryId, sev cdb.JsonKV) error {
	qs := func(tx *sql.Tx) (err error) {
		var ret int
		err = tx.QueryRow("SELECT entry_set_done($1, $2)", id.String(), sev).Scan(&ret)
		if err == nil {
			log.Printf("entry set done ret: %v\n", ret)
		} else {
			logger().Warnw("setDone fail", "err", err)
		}
		return
	}
	return mw.withTxQuery(qs)
}

func (mw *MetaWrap) Save(entry *Entry, is_update bool) error {
	var qs func(tx *sql.Tx) (err error)
	if is_update {
		qs = func(tx *sql.Tx) (err error) {
			query := "UPDATE " + mw.table() + " SET app_id = $1, author = $2 WHERE id = $3"
			var r sql.Result
			r, err = tx.Exec(query, entry.AppId, entry.Author, entry.Id)
			if err == nil {
				a, _ := r.RowsAffected()
				log.Printf("entry '%s' updated: %v", entry.Id, a)
			} else {
				logger().Warnw("save fail", "err", err)
			}
			return
		}
	} else {
		qs = func(tx *sql.Tx) (err error) {
			query := "SELECT entry_save($1, $2, $3, $4, $5, $6, $7, $8, $9);"

			err = tx.QueryRow(query, mw.table_suffix,
				entry.Id.String(), entry.Path, entry.Meta, entry.sev, entry.Hashes, entry.Ids,
				entry.AppId, entry.Author).Scan(&entry.ret)
			if err == nil {
				log.Printf("entry save ret: %v\n", entry.ret)
			}
			return
		}
	}

	return mw.withTxQuery(qs)
}

func (mw *MetaWrap) BatchSave(entries []*Entry) error {
	qs := func(tx *sql.Tx) error {

		sql := "SELECT entry_save($1, $2, $3, $4, $5, $6, $7, $8, $9);"
		st, err := tx.Prepare(sql)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			err := st.QueryRow(mw.table_suffix,
				entry.Id.String(), entry.Path, entry.Meta, entry.sev, entry.Hashes, entry.Ids,
				entry.AppId, entry.Author).Scan(&entry.ret)
			if err != nil {
				log.Printf("batchSave %s %s error: %s", entry.Id.String(), entry.Path, err)
				return err
			}
			log.Printf("batchSave %s %s %d: %d", entry.Path, entry.Mime, entry.Size, entry.ret)
		}
		return nil
	}

	return mw.withTxQuery(qs)
}

type ehash struct {
	hash, path, id string
}

func (mw *MetaWrap) GetHash(hash string) (*ehash, error) {
	db := mw.getDb()
	var e = ehash{hash: hash}
	sql := "SELECT item_id, path FROM " + tableHash(hash) + " WHERE hashed = $1 LIMIT 1"
	row := db.QueryRow(sql, hash)
	err := row.Scan(&e.id, &e.path)
	if err != nil {
		logger().Warnw("get hash fail", "hash", hash)
		return nil, err
	}
	return &e, nil
}

func (mw *MetaWrap) GetEntry(id EntryId) (*Entry, error) {
	db := mw.getDb()
	sql := "SELECT name, path, mime, size, sev, status, created, roofs FROM " + tableMap(id.String()) + " WHERE id = $1 LIMIT 1"
	row := db.QueryRow(sql, id.String())
	var e = Entry{Id: &id}
	err := row.Scan(&e.Name, &e.Path, &e.Mime, &e.Size, &e.sev, &e.Status, &e.Created, &e.Roofs)
	if err != nil {
		log.Printf("[%s]getEntry %s error %s", mw.roof, id.String(), err)
		return nil, err
	}
	return &e, nil
}

func (mw *MetaWrap) Delete(id EntryId) error {
	qs := func(tx *sql.Tx) (err error) {
		var ret int
		sql := "SELECT entry_delete($1, $2);"
		err = tx.QueryRow(sql, mw.table_suffix, id.String()).Scan(&ret)
		if err == nil {
			log.Printf("delete entry [%s]%s result %v", mw.table_suffix, id.String(), ret)
		}
		return
	}
	return mw.withTxQuery(qs)
}

func (mw *MetaWrap) MapTags(id EntryId, tags string) error {

	var qtags, err = cdb.NewQarrayText(tags)
	if err != nil {
		return err
	}
	qs := func(tx *sql.Tx) (err error) {
		var ret int
		sql := "SELECT tag_map($1, $2, $3);"
		err = tx.QueryRow(sql, mw.table_suffix, id, qtags).Scan(&ret)
		if err == nil {
			log.Printf("entry [%s]%v mapping tags '%s' result %v", mw.table_suffix, id, tags, ret)
		}
		return
	}
	return mw.withTxQuery(qs)
}

func (mw *MetaWrap) UnmapTags(id EntryId, tags string) error {

	var qtags, err = cdb.NewQarrayText(tags)
	if err != nil {
		return err
	}
	qs := func(tx *sql.Tx) (err error) {
		var ret int
		sql := "SELECT tag_unmap($1, $2, $3);"
		err = tx.QueryRow(sql, mw.table_suffix, id, qtags).Scan(&ret)
		if err == nil {
			log.Printf("entry [%s]%v unmap tags '%s' result %v", mw.table_suffix, id, tags, ret)
		}
		return
	}
	return mw.withTxQuery(qs)
}

func (mw *MetaWrap) withTxQuery(query func(tx *sql.Tx) error) error {
	return withTxQuery(mw.roof, query)
}

func (mw *MetaWrap) getDb() *sql.DB {
	return getDb(mw.roof)
}

func tableHash(s string) string {
	return hash_table_prefix + s[:1]
}

func tableMap(s string) string {
	return map_table_prefix + s[:1]
}
