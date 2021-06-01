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

	"github.com/go-imsto/imagid"
	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/image"
	cdb "github.com/go-imsto/imsto/storage/types"
)

const (
	prefixHashTable = "hash_"
	prefixMapTable  = "mapping_"
	prefixMetaTable = "meta_"
	maxArgs         = 10
	commonRoof      = "common"
)

// HashEntry ...
type HashEntry struct {
	ID   IID    `json:"id"`
	Path string `json:"path"`
}

// MetaWrap ...
type MetaWrap struct {
	roof string
	// prefix  string

	tableSuffix string
}

func newMetaWrap(roof string) *MetaWrap {
	mw := &MetaWrap{roof: roof, tableSuffix: roof}

	return mw
}

var (
	metaWrappers   = make(map[string]MetaWrapper)
	metaColumns    = "id, path, name, meta, hashes, ids, size, sev, tags, exif, app_id, author, created, roof"
	sortableFields = []string{"id", "created"}
	ErrDbError     = errors.New("database error")
)

const (
	metaCreateTmpl = `CREATE TABLE IF NOT EXISTS meta_%s ( LIKE meta_template INCLUDING ALL )`
)

func InitMetaTables() {
	db := getDb()
	roofs := config.Current.Engines
	logger().Infow("checking or create tables of metas", "roofs", len(roofs))
	for k := range roofs {
		_, err := db.Exec(fmt.Sprintf(metaCreateTmpl, k))
		if err != nil {
			logger().Fatalw("create table of meta_? fail", "roof", k, "err", err)
		}
	}
}

// NewMetaWrapper ...
func NewMetaWrapper(roof string) (mw MetaWrapper) {
	if roof == "" {
		panic(ErrEmptyRoof)
	}
	var ok bool
	if mw, ok = metaWrappers[roof]; !ok {
		mw = newMetaWrap(roof)
		metaWrappers[roof] = mw
	}

	return mw
}

func (mw *MetaWrap) table() string {
	return prefixMetaTable + mw.tableSuffix
}

const (
	ASCENDING  = 1
	DESCENDING = -1
)

func isSortable(k string) bool {
	for _, sf := range sortableFields {
		if k == sf {
			return true
		}
	}
	return false
}

func buildWhere(filter MetaFilter) (where string, args []interface{}) {
	where = " WHERE status = 0"

	argc := 0
	args = make([]interface{}, 0, maxArgs)

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

// Count ...
func (mw *MetaWrap) Count(filter MetaFilter) (t int, err error) {
	db := mw.getDb()
	table := mw.table()

	t = 0
	where, args := buildWhere(filter)
	// err = db.QueryRow("SELECT COUNT(id) FROM "+table+where, args...).Scan(&t)
	rows, err := db.Query("SELECT COUNT(id) FROM "+table+where, args...)
	// return &Row{rows: rows, err: err}
	if err != nil {
		logger().Warnw("query count fail", "table", table, "err", err)
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
	str := "SELECT " + metaColumns + " FROM " + table + where

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
			logger().Infow("bind fail", "err", err)
			return
		}
		a = append(a, entry)
	}
	return
}

// depends metaColumns
func _bindRow(rs rowScanner) (*Entry, error) {

	e := Entry{}
	var id, roof string
	var meta image.Attr
	// "id, path, name, meta, hashes, ids, size, sev, exif, app_id, author, created, roof"
	err := rs.Scan(&id, &e.Path, &e.Name, &meta, &e.Hashes, &e.IDs, &e.Size,
		&e.sev, &e.Tags, &e.exif, &e.AppId, &e.Author, &e.Created, &roof)
	if err != nil {
		logger().Infow("bind fail", "err", err)
		err = ErrDbError
		return nil, err
	}
	e.Id, err = imagid.ParseID(id)
	if err != nil {
		return nil, err
	}

	e.Meta = &meta
	e.Roofs = cdb.StringArray{roof}

	return &e, nil
}

// NextID ...
func (mw *MetaWrap) NextID() (nid uint64, err error) {
	err = mw.getDb().QueryRow("SELECT shard_1.id_generator() as id").Scan(&nid)
	return
}

func (mw *MetaWrap) GetMeta(id string) (entry *Entry, err error) {
	db := mw.getDb()

	sql := "SELECT " + metaColumns + " FROM " + mw.table() + " WHERE id = $1 LIMIT 1"

	row := db.QueryRow(sql, id)

	entry, err = _bindRow(row)
	if err != nil {
		logger().Infow("bind fail", "id", id, "err", err)
		return
	}

	return
}

func popPrepared() (*Entry, error) {
	mwr := NewMetaWrapper(commonRoof)
	mw := mwr.(*MetaWrap)

	db := mw.getDb()

	sql := "SELECT " + metaColumns + " FROM meta__prepared ORDER BY created ASC LIMIT 1"

	row := db.QueryRow(sql)

	return _bindRow(row)
}

func (mw *MetaWrap) Ready(entry *Entry) error {
	return mw.withTxQuery(func(tx *sql.Tx) error {
		var created time.Time
		err := tx.QueryRow("SELECT created FROM meta__prepared WHERE id = $1", entry.Id).Scan(&created)
		if err == nil {
			logger().Infow("check prepared with id exist", "id", entry.Id, "created", created.Format(time.RFC3339))
			return nil
		}
		logger().Infow("check prepared with id not exist", "id", entry.Id, "err", err)

		err = tx.QueryRow("SELECT created FROM meta__prepared WHERE hashes->>'hash' = $1", entry.GetHash()).Scan(&created)
		if err == nil {
			logger().Infow("check prepared with hash exist", "hash", entry.GetHash(), "created", created.Format(time.RFC3339))
			return nil
		}
		logger().Infow("check prepared with hash not exist", "hash", entry.GetHash(), "err", err)

		_, err = tx.Exec(`INSERT INTO meta__prepared (id, roof, path, name, size, meta, hashes, ids, app_id, author, tags)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`, entry.Id, mw.tableSuffix, entry.Path,
			entry.Name, entry.Size, entry.Meta, entry.Hashes, entry.IDs,
			entry.AppId, entry.Author, entry.Tags)
		if err != nil {
			logger().Warnw("save prepared fail", "entry", entry, "err", err)
		} else {
			logger().Infow("save prepared OK", "id", entry.Id, "hashs", entry.Hashes)
		}
		return err
	})
}

func (mw *MetaWrap) SetDone(id string, sev cdb.Meta) error {
	qs := func(tx *sql.Tx) (err error) {
		var ret int
		err = tx.QueryRow("SELECT entry_set_done($1, $2)", id, sev).Scan(&ret)
		if err == nil {
			logger().Infow("setDone", "id", id)
		} else {
			logger().Warnw("setDone fail", "id", id, "err", err)
		}
		return
	}
	return mw.withTxQuery(qs)
}

// Save ...
func (mw *MetaWrap) Save(entry *Entry, isUpdate bool) error {
	var qs func(tx *sql.Tx) (err error)
	if isUpdate {
		qs = func(tx *sql.Tx) (err error) {
			query := "UPDATE " + mw.table() + " SET app_id = $1, author = $2 WHERE id = $3"
			var r sql.Result
			r, err = tx.Exec(query, entry.AppId, entry.Author, entry.Id)
			if err == nil {
				a, _ := r.RowsAffected()
				logger().Infow("entry updated", "id", entry.Id, "ra", a)
			} else {
				logger().Warnw("save fail", "err", err)
			}
			return
		}
	} else {
		qs = func(tx *sql.Tx) (err error) {
			query := "SELECT entry_save($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);"
			err = tx.QueryRow(query, mw.tableSuffix,
				entry.Id, entry.Path, entry.Name, entry.Size, entry.Meta, entry.sev, entry.Hashes, entry.IDs,
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

		sql := "SELECT entry_save($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);"
		st, err := tx.Prepare(sql)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			err := st.QueryRow(mw.tableSuffix,
				entry.Id, entry.Path, entry.Name, entry.Size, entry.Meta, entry.sev, entry.Hashes, entry.IDs,
				entry.AppId, entry.Author).Scan(&entry.ret)
			if err != nil {
				log.Printf("batchSave %s %s error: %s", entry.Id, entry.Path, err)
				return err
			}
			log.Printf("batchSave %s %d: %d", entry.Path, entry.Size, entry.ret)
		}
		return nil
	}

	return mw.withTxQuery(qs)
}

func (mw *MetaWrap) GetHash(hash string) (he *HashEntry, err error) {
	db := mw.getDb()
	he = new(HashEntry)
	q := "SELECT item_id, path FROM " + tableHash(hash) + " WHERE hashed = $1"
	err = db.QueryRow(q, hash).Scan(&he.ID, &he.Path)
	if err != nil {
		if err == sql.ErrNoRows {
			logger().Infow("row not found", "hash", hash)
		} else {
			logger().Warnw("get hash fail", "hash", hash)
		}

	}
	return
}

func (mw *MetaWrap) GetMapping(id string) (*mapItem, error) {
	db := mw.getDb()
	sql := "SELECT name, path, size, sev, status, created, roofs FROM " + tableMap(id) + " WHERE id = $1 LIMIT 1"
	row := db.QueryRow(sql, id)
	iID, _ := imagid.ParseID(id)
	var e = mapItem{ID: iID}
	err := row.Scan(&e.Name, &e.Path, &e.Size, &e.sev, &e.Status, &e.Created, &e.Roofs)
	if err != nil {
		logger().Infow("GetMapping fail", "roof", mw.roof, "id", id, "err", err)
		return nil, err
	}
	return &e, nil
}

func (mw *MetaWrap) Delete(id string) error {
	qs := func(tx *sql.Tx) (err error) {
		var ret int
		sql := "SELECT entry_delete($1, $2);"
		err = tx.QueryRow(sql, mw.tableSuffix, id).Scan(&ret)
		if err == nil {
			log.Printf("delete entry [%s]%s result %v", mw.tableSuffix, id, ret)
		}
		return
	}
	return mw.withTxQuery(qs)
}

func (mw *MetaWrap) MapTags(id string, tags string) error {

	var qtags, err = cdb.NewQarrayText(tags)
	if err != nil {
		return err
	}
	qs := func(tx *sql.Tx) (err error) {
		var ret int
		sql := "SELECT tag_map($1, $2, $3);"
		err = tx.QueryRow(sql, mw.tableSuffix, id, qtags).Scan(&ret)
		if err == nil {
			log.Printf("entry [%s]%v mapping tags '%s' result %v", mw.tableSuffix, id, tags, ret)
		}
		return
	}
	return mw.withTxQuery(qs)
}

func (mw *MetaWrap) UnmapTags(id string, tags string) error {

	var qtags, err = cdb.NewQarrayText(tags)
	if err != nil {
		return err
	}
	qs := func(tx *sql.Tx) (err error) {
		var ret int
		sql := "SELECT tag_unmap($1, $2, $3);"
		err = tx.QueryRow(sql, mw.tableSuffix, id, qtags).Scan(&ret)
		if err == nil {
			log.Printf("entry [%s]%v unmap tags '%s' result %v", mw.tableSuffix, id, tags, ret)
		}
		return
	}
	return mw.withTxQuery(qs)
}

func (mw *MetaWrap) withTxQuery(query func(tx *sql.Tx) error) error {
	return withTxQuery(query)
}

func (mw *MetaWrap) getDb() *sql.DB {
	return getDb()
}

func tableHash(s string) string {
	return prefixHashTable + s[:1]
}

func tableMap(s string) string {
	return prefixMapTable + s[:2]
}
