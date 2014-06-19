package storage

import (
	"database/sql"
	_ "database/sql/driver"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"strings"
	"wpst.me/calf/config"
	cdb "wpst.me/calf/db"
	"wpst.me/calf/image"
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
	SetDone(id EntryId, sev cdb.Hstore) error
	Save(entry *Entry) error
	BatchSave(entries []*Entry) error
	GetMeta(id EntryId) (*Entry, error)
	GetHash(hash string) (*ehash, error)
	GetEntry(id EntryId) (*Entry, error)
	Delete(id EntryId) error
}

type rowScanner interface {
	Scan(...interface{}) error
}

type MetaWrap struct {
	dsn  string
	roof string
	// prefix  string

	table_suffix string
}

func newMetaWrap(roof string) *MetaWrap {
	if roof == "" {
		log.Print("arg roof is empty")
		roof = "common"
	}
	dsn := config.GetValue(roof, "meta_dsn")
	// log.Printf("[%s]dsn: %s", roof, dsn)
	table := config.GetValue(roof, "meta_table_suffix")
	if table == "" {
		// log.Print("meta_table_suffix is empty, use roof")
		table = roof
	}
	log.Printf("table suffix: %s", table)
	mw := &MetaWrap{dsn: dsn, roof: roof, table_suffix: table}

	return mw
}

var meta_wrappers = make(map[string]MetaWrapper)

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

var (
	meta_columns    = "id, path, name, meta, hashes, ids, size, sev, tags, exif, app_id, author, created, roof"
	sortable_fields = []string{"id", "created"}
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
	Tags string
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

	return
}

func (mw *MetaWrap) Count(filter MetaFilter) (t int, err error) {
	db := mw.getDb()
	defer db.Close()
	table := mw.table()

	t = 0
	where, args := buildWhere(filter)
	// err = db.QueryRow("SELECT COUNT(id) FROM "+table+where, args...).Scan(&t)
	rows, err := db.Query("SELECT COUNT(id) FROM "+table+where, args...)
	// return &Row{rows: rows, err: err}
	if err != nil {
		log.Printf("query count error: %s", err)
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
	defer db.Close()

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
	var meta cdb.Hstore
	// "id, path, name, meta, hashes, ids, size, sev, exif, app_id, author, created, roof"
	err := rs.Scan(&id, &e.Path, &e.Name, &meta, &e.Hashes, &e.Ids, &e.Size,
		&e.sev, &e.Tags, &e.exif, &e.AppId, &e.Author, &e.Created, &roof)
	if err != nil {
		return nil, err
	}
	e.Id, err = NewEntryId(id)
	if err != nil {
		return nil, err
	}

	var ia image.Attr
	err = meta.ToStruct(&ia)
	if err != nil {
		return nil, err
	}

	// log.Printf("bind meta %s: %v", meta, ia)

	e.Meta = &ia
	e.Mime = fmt.Sprint(meta.Get("mime"))
	e.Roofs = cdb.Qarray{roof}

	// log.Printf("bind name: %s, path: %s, size: %d, mime: %s\n", e.Name, e.Path, e.Size, e.Mime)

	return &e, nil
}

func (mw *MetaWrap) GetMeta(id EntryId) (*Entry, error) {
	db := mw.getDb()
	defer db.Close()

	sql := "SELECT " + meta_columns + " FROM " + mw.table() + " WHERE id = $1 LIMIT 1"

	row := db.QueryRow(sql, id.String())

	return _bindRow(row)
}

func popPrepared() (*Entry, error) {
	mwr := NewMetaWrapper("common")
	mw := mwr.(*MetaWrap)

	db := mw.getDb()
	defer db.Close()

	sql := "SELECT " + meta_columns + " FROM meta__prepared ORDER BY created ASC LIMIT 1"

	row := db.QueryRow(sql)

	return _bindRow(row)
}

func (mw *MetaWrap) Ready(entry *Entry) error {
	db := mw.getDb()
	defer db.Close()

	sql := "SELECT entry_ready($1, $2, $3, $4, $5, $6, $7, $8, $9);"
	row := db.QueryRow(sql, mw.table_suffix,
		entry.Id.String(), entry.Path, entry.Meta.Hstore(), entry.Hashes, entry.Ids,
		entry.AppId, entry.Author, entry.Tags)

	var ret int
	err := row.Scan(&ret)
	if err != nil {
		return err
	}

	log.Printf("entry ready ret: %v\n", ret)

	return nil
}

func (mw *MetaWrap) SetDone(id EntryId, sev cdb.Hstore) error {
	db := mw.getDb()
	defer db.Close()
	var ret int
	err := db.QueryRow("SELECT entry_set_done($1, $2)", id.String(), sev).Scan(&ret)
	if err != nil {
		return err
	}

	log.Printf("entry set done ret: %v\n", ret)

	return nil
}

func (mw *MetaWrap) Save(entry *Entry) error {
	db := mw.getDb()
	defer db.Close()

	// log.Printf("hashes: %s\n", entry.Hashes)
	// log.Printf("ids: %s\n", entry.Ids)

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	sql := "SELECT entry_save($1, $2, $3, $4, $5, $6, $7, $8, $9);"

	err = tx.QueryRow(sql, mw.table_suffix,
		entry.Id.String(), entry.Path, entry.Meta.Hstore(), entry.sev, entry.Hashes, entry.Ids,
		entry.AppId, entry.Author).Scan(&entry.ret)

	if err != nil {
		tx.Rollback()
		return err
	}

	log.Printf("entry save ret: %v\n", entry.ret)

	tx.Commit()
	return nil
}

func (mw *MetaWrap) BatchSave(entries []*Entry) error {
	db := mw.getDb()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	sql := "SELECT entry_save($1, $2, $3, $4, $5, $6, $7, $8, $9);"
	st, err := tx.Prepare(sql)
	for _, entry := range entries {
		err := st.QueryRow(mw.table_suffix,
			entry.Id.String(), entry.Path, entry.Meta.Hstore(), entry.sev, entry.Hashes, entry.Ids,
			entry.AppId, entry.Author).Scan(&entry.ret)
		if err != nil {
			log.Printf("batchSave %s %s error: %s", entry.Id.String(), entry.Path, err)
			tx.Rollback()
			return err
		}
		log.Printf("batchSave %s %s %d: %d", entry.Path, entry.Mime, entry.Size, entry.ret)
	}
	tx.Commit()
	return nil
}

type ehash struct {
	hash, path, id string
}

func (mw *MetaWrap) GetHash(hash string) (*ehash, error) {
	db := mw.getDb()
	defer db.Close()
	var e = ehash{hash: hash}
	sql := "SELECT item_id, path FROM " + tableHash(hash) + " WHERE hashed = $1 LIMIT 1"
	row := db.QueryRow(sql, hash)
	err := row.Scan(&e.id, &e.path)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &e, nil
}

func (mw *MetaWrap) GetEntry(id EntryId) (*Entry, error) {
	db := mw.getDb()
	defer db.Close()
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
	db := mw.getDb()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	sql := "SELECT entry_delete($1, $2);"
	var ret int
	err = tx.QueryRow(sql, mw.table_suffix, id.String()).Scan(&ret)

	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	log.Printf("delete entry [%s]%s result %v", mw.table_suffix, id.String(), ret)
	return nil
}

func (mw *MetaWrap) getDb() *sql.DB {
	db, err := sql.Open("postgres", mw.dsn)

	if err != nil {
		log.Fatalf("open db error: %s", err)
	}
	return db
}

func tableHash(s string) string {
	return hash_table_prefix + s[:1]
}

func tableMap(s string) string {
	return map_table_prefix + s[:1]
}
