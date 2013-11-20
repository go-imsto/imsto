package storage

import (
	"calf/config"
	"database/sql"
	_ "database/sql/driver"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

type Ticket struct {
	section string
	dsn     string
	table   string
	appid   AppId
	author  Author
	prompt  string
	id      int
}

func newTicket(section string, appid AppId) *Ticket {
	dsn := config.GetValue(section, "meta_dsn")
	table := config.GetValue(section, "ticket_table")
	if table == "" {
		log.Print("'ticket_table' is not found in config")
	}
	// log.Printf("table: %s", table)
	t := &Ticket{section: section, dsn: dsn, table: table, appid: appid}

	return t
}

func NewTokenRequest(r *http.Request) (t *apiToken, err error) {
	var (
		section string
		appid   AppId
		author  Author
	)

	if err = r.ParseForm(); err != nil {
		log.Print("form parse error:", err)
		return
	}

	section, appid, author, err = parseRequest(r)
	if err != nil {
		log.Print("request error:", err)
		return
	}

	t, err = getApiToken(section, appid)
	if err != nil {
		return
	}

	t.SetValue([]byte(fmt.Sprint(author)), VC_TOKEN)
	return
}

func NewTicketRequest(r *http.Request) (t *apiToken, err error) {
	var (
		section string
		appid   AppId
		author  Author
	)
	if err = r.ParseForm(); err != nil {
		log.Print("form parse error:", err)
		return
	}

	section, appid, author, err = parseRequest(r)
	if err != nil {
		log.Print("request error:", err)
		return
	}

	t, err = getApiToken(section, appid)
	if err != nil {
		return
	}
	var ok bool
	ok, err = t.VerifyString(r.PostFormValue("token"))
	if err != nil {
		return
	}
	if !ok {
		err = errors.New("Invalid Token")
	}
	ticket := newTicket(section, appid)

	ticket.author = author
	ticket.prompt = r.PostFormValue("prompt")

	err = ticket.saveNew()

	if err != nil {
		log.Printf("save ticket error %s", err)
		return
	}

	t, err = getApiToken(section, appid)
	if err != nil {
		return
	}
	t.SetValue([]byte(fmt.Sprint(ticket.GetId())), VC_TICKET)
	log.Printf("new token: %x", t.Binary())

	return
}

func (t *Ticket) saveNew() error {
	db := t.getDb()
	defer db.Close()

	var id int
	sql := "INSERT INTO " + t.table + "(app_id, author, prompt) VALUES($1, $2, $3) RETURNING id"
	log.Printf("sql: %s", sql)
	err := db.QueryRow(sql, t.appid, t.author, t.prompt).Scan(&id)
	if err != nil {
		return err
	}
	t.id = id
	return nil
}

func (t *Ticket) update() error {
	// TODO:
	return nil
}

func (t *Ticket) load(id int) error {
	db := t.getDb()
	defer db.Close()
	t.id = id
	sql := "SELECT app_id, author, prompt FROM " + t.table + " WHERE id = $1 LIMIT 1"
	err := db.QueryRow(sql, t.id).Scan(&t.appid, &t.author, &t.prompt)
	if err != nil {
		return err
	}
	log.Printf("ticket #%d loaded", id)
	return nil
}

func (t *Ticket) bindEntry(entry *Entry) error {
	db := t.getDb()
	defer db.Close()
	log.Printf("start binding %s", entry.Id)
	// sql := "UPDATE " + t.table + " SET img_id=$1, img_path=$1, img_meta=$2, uploaded=$3, updated=$4 WHERE id = $5"
	// sr, err := db.Exec(sql, entry.Id.String(), entry.Path, entry.Meta.Hstore(), true, entry.Created, t.id)
	// log.Printf("db exec result: %s, error:", sr, err)
	// if err != nil {
	// 	return err
	// }
	// var ra int64
	// ra, err = sr.RowsAffected()
	sql := "SELECT ticket_update($1, $2, $3)"
	var ret int
	mw := NewMetaWrapper(t.section)
	err := db.QueryRow(sql, t.id, mw.TableSuffix(), entry.Id.String()).Scan(&ret)
	if err != nil {
		return err
	}
	log.Printf("update result: %d", ret)
	return nil
}

func (t *Ticket) GetId() int {
	return t.id
}

func (t *Ticket) getDb() *sql.DB {
	db, err := sql.Open("postgres", t.dsn)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
