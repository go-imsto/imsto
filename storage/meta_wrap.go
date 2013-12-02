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
)

type MetaWrapper interface {
	Browse(limit, offset int, sort map[string]int) ([]Entry, int, error)
	Save(entry *Entry) error
	BatchSave(entries []*Entry) error
	GetMeta(id EntryId) (*Entry, error)
	GetHash(hash string) (*ehash, error)
	GetEntry(id EntryId) (*Entry, error)
	Delete(id EntryId) error
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
	valid_condition = " WHERE status = 0"
)

const (
	ASCENDING  = 1
	DESCENDING = -1
)

var (
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

func (mw *MetaWrap) Browse(limit, offset int, sort map[string]int) (a []Entry, t int, err error) {
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

	t = 0
	err = db.QueryRow("SELECT COUNT(id) FROM " + table + valid_condition).Scan(&t)
	if err != nil {
		log.Printf("query count error: %s", err)
		return
	}

	if t == 0 {
		log.Print("browse empty meta")
		return
	}

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
	str := "SELECT id, name, path, size, meta, sev, created, ids, hashes FROM " + table + valid_condition

	if len(orders) > 0 {
		str = str + " ORDER BY " + strings.Join(orders, ",")
	}
	// log.Printf("sql: %s", str)
	var r *sql.Rows
	r, err = db.Query(str+" LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return
	}

	defer r.Close()

	for r.Next() {
		e := Entry{}
		var id string
		var meta, sev cdb.Hstore
		err = r.Scan(&id, &e.Name, &e.Path, &e.Size, &meta, &sev, &e.Created, &e.Ids, &e.Hashes)
		if err != nil {
			return
		}
		e.Id, err = NewEntryId(id)
		if err != nil {
			return
		}

		var ia image.Attr
		err = meta.ToStruct(&ia)
		if err != nil {
			return
		}

		e.Meta = &ia
		e.Mime = fmt.Sprint(meta.Get("mime"))

		a = append(a, e)
	}
	return
}

func (mw *MetaWrap) GetMeta(id EntryId) (*Entry, error) {
	db := mw.getDb()
	defer db.Close()

	sql := "SELECT name, path, size, meta, sev, ids, hashes FROM " + mw.table() + " WHERE id = $1 LIMIT 1"
	entry := Entry{Id: &id}
	row := db.QueryRow(sql, id.String())

	var meta cdb.Hstore
	err := row.Scan(&entry.Name, &entry.Path, &entry.Size, &meta, &entry.sev, &entry.Ids, &entry.Hashes)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println("first id:", entry.Ids[0])

	log.Println("meta:", meta)
	log.Println("sev:", entry.sev)
	var ia image.Attr
	err = meta.ToStruct(&ia)
	if err != nil {
		log.Println(err)
		return &entry, err
	}
	log.Println("ia:", ia)

	entry.Meta = &ia
	entry.Mime = fmt.Sprint(meta.Get("mime"))

	log.Printf("name: %s, path: %s, size: %d, mime: %s\n", entry.Name, entry.Path, entry.Size, entry.Mime)

	return &entry, nil
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

	sql := "SELECT entry_save($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);"
	row := tx.QueryRow(sql, mw.table_suffix,
		entry.Id.String(), entry.Path, entry.Name, entry.Mime, entry.Size, entry.Meta.Hstore(), entry.sev, entry.Hashes, entry.Ids)

	var ret int
	err = row.Scan(&ret)

	if err != nil {
		tx.Rollback()
		return err
		// log.Fatal(err)
	}

	log.Printf("entry save ret: %v\n", ret)

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

	sql := "SELECT entry_save($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);"
	st, err := tx.Prepare(sql)
	for _, entry := range entries {
		var ret int
		err := st.QueryRow(mw.table_suffix,
			entry.Id.String(), entry.Path, entry.Name, entry.Mime, entry.Size, entry.Meta.Hstore(), entry.sev, entry.Hashes, entry.Ids).Scan(&ret)
		if err != nil {
			log.Printf("batchSave %s %s error: %s", entry.Id.String(), entry.Path, err)
			tx.Rollback()
			return err
		}
		log.Printf("batchSave %s %s %d: %d", entry.Path, entry.Mime, entry.Size, ret)
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
	sql := "SELECT name, path, mime, size, sev, status, created FROM " + tableMap(id.String()) + " WHERE id = $1 LIMIT 1"
	row := db.QueryRow(sql, id.String())
	var e = Entry{Id: &id}
	err := row.Scan(&e.Name, &e.Path, &e.Mime, &e.Size, &e.sev, &e.Status, &e.Created)
	if err != nil {
		log.Printf("[%s]getEntry %s error %s", mw.roof, id.String(), err)
		return nil, err
	}
	return &e, nil
}

func (mw *MetaWrap) Delete(id EntryId) error {
	db := mw.getDb()
	defer db.Close()
	sql := "UPDATE " + mw.table() + " SET status = 1 WHERE id = $1"
	r, err := db.Exec(sql, id.String())
	if err != nil {
		log.Println(err)
		return err
	}
	var af int64
	af, err = r.RowsAffected()
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("delete entry %s result %v", id.String(), af)
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
