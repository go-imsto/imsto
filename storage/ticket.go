package storage

import (
	"database/sql"
	_ "database/sql/driver"
	"encoding/binary"
	"errors"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"wpst.me/calf/config"
)

type Ticket struct {
	roof    string
	dsn     string
	table   string
	AppId   AppId  `json:"appid,omitempty"`
	Author  Author `json:"author,omitempty"`
	Prompt  string `json:"prompt,omitempty"`
	id      int
	ImgId   string `json:"img_id,omitempty"`
	ImgPath string `json:"img_path,omitempty"`
	Done    bool   `json:"done,omitempty"`
}

func newTicket(roof string, appid AppId) *Ticket {
	dsn := config.GetValue(roof, "meta_dsn")
	table := getTicketTable(roof)
	// log.Printf("table: %s", table)
	t := &Ticket{roof: roof, dsn: dsn, table: table, AppId: appid}

	return t
}

func TokenRequestNew(r *http.Request) (t *apiToken, err error) {
	var (
		roof   string
		appid  AppId
		author Author
	)
	roof, appid, author, err = parseRequest(r)
	if err != nil {
		log.Print("request error:", err)
		return
	}

	t, err = getApiToken(roof, appid)
	if err != nil {
		return
	}
	var b = make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(author))
	t.SetValue(b, VC_TOKEN)
	return
}

func TicketRequestNew(r *http.Request) (t *apiToken, err error) {
	var (
		roof   string
		appid  AppId
		author Author
	)
	roof, appid, author, err = parseRequest(r)
	if err != nil {
		log.Print("request error:", err)
		return
	}

	t, err = getApiToken(roof, appid)
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
	ticket := newTicket(roof, appid)

	ticket.Author = author
	ticket.Prompt = r.FormValue("prompt")

	err = ticket.saveNew()

	if err != nil {
		log.Printf("save ticket error %s", err)
		return
	}

	t, err = getApiToken(roof, appid)
	if err != nil {
		return
	}

	var b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(ticket.GetId()))
	t.SetValue(b, VC_TICKET)
	log.Printf("new token: %x", t.Binary())

	return
}

func TicketRequestLoad(r *http.Request) (ticket *Ticket, err error) {
	var (
		roof   string
		appid  AppId
		author Author
	)
	roof, appid, author, err = parseRequest(r)
	if err != nil {
		log.Print("request error:", err)
		return
	}

	var t *apiToken
	t, err = getApiToken(roof, appid)
	if err != nil {
		return
	}
	var ok bool
	ok, err = t.VerifyString(r.FormValue("token"))
	if err != nil {
		return
	}
	if !ok {
		err = errors.New("Invalid Token")
	}

	id := t.GetValuleInt() // int(binary.BigEndian.Uint64(t.GetValue()))

	ticket, err = loadTicket(roof, int(id))
	if ticket.Author != author {
		log.Printf("mismatch author %s : %s", ticket.Author, author)
	}
	return
}

func (t *Ticket) saveNew() error {
	db := t.getDb()
	defer db.Close()

	var id int
	sql := "INSERT INTO " + t.table + "(roof, app_id, author, prompt) VALUES($1, $2, $3, $4) RETURNING id"
	log.Printf("sql: %s", sql)
	err := db.QueryRow(sql, t.roof, t.AppId, t.Author, t.Prompt).Scan(&id)
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
	sql := "SELECT roof, app_id, author, prompt, img_id, img_path, done FROM " + t.table + " WHERE id = $1 LIMIT 1"
	err := db.QueryRow(sql, id).Scan(&t.roof, &t.AppId, &t.Author, &t.Prompt, &t.ImgId, &t.ImgPath, &t.Done)
	if err != nil {
		return err
	}
	log.Printf("ticket #%d loaded", id)
	return nil
}

func loadTicket(sn string, id int) (t *Ticket, err error) {
	t = newTicket(sn, AppId(0))
	err = t.load(id)
	return
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
	sql := "SELECT ticket_update($1, $2)"
	var ret int
	err := db.QueryRow(sql, t.id, entry.Id.String()).Scan(&ret)
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

func getTicketTable(sn string) (table string) {
	table = config.GetValue(sn, "ticket_table")
	if table == "" {
		log.Print("'ticket_table' is not found in config")
	}
	return
}
