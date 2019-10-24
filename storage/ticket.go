package storage

import (
	"database/sql"
	"encoding/binary"
	"log"
)

type Ticket struct {
	roof    string
	table   string
	AppID   AppID  `json:"appid,omitempty"`
	Author  Author `json:"author,omitempty"`
	Prompt  string `json:"prompt,omitempty"`
	id      int
	ImgId   string `json:"img_id,omitempty"`
	ImgPath string `json:"img_path,omitempty"`
	Done    bool   `json:"done,omitempty"`
}

func newTicket(roof string, appid AppID) *Ticket {
	t := &Ticket{roof: roof, table: "upload_ticket", AppID: appid}

	return t
}

// TicketRequestNew ... deprecated
func (a *App) TicketRequestNew(roof, token string, uid int, prompt string) (t *apiToken, err error) {
	ticket := newTicket(roof, a.Id)

	ticket.Author = Author(uid)
	ticket.Prompt = prompt

	err = ticket.saveNew()

	if err != nil {
		logger().Warnw("ticket save fail", "err", err)
		return
	}

	t, err = a.VerifyToken(token)
	if err != nil {
		logger().Warnw("genToken fail", "err", err)
		return
	}

	var b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(ticket.GetId()))
	t.SetValue(b, VC_TICKET)
	log.Printf("new token: %x", t.Binary())

	return
}

func (a *App) TicketRequestLoad(token, roof string, author int) (ticket *Ticket, err error) {

	at, err := a.VerifyToken(token)
	if err != nil {
		return
	}

	id := at.GetValuleInt() // int(binary.BigEndian.Uint64(t.GetValue()))

	ticket, err = loadTicket(roof, int(id))
	if ticket.Author != Author(author) {
		log.Printf("mismatch author %d : %d", ticket.Author, author)
	}
	return
}

func (t *Ticket) saveNew() error {
	db := t.getDb()
	defer db.Close()

	var id int
	sql := "INSERT INTO " + t.table + "(roof, app_id, author, prompt) VALUES($1, $2, $3, $4) RETURNING id"
	log.Printf("save ticket sql: %s", sql)
	err := db.QueryRow(sql, t.roof, t.AppID, t.Author, t.Prompt).Scan(&id)
	if err != nil {
		return err
	}
	t.id = id
	log.Printf("new ticket id: %4d", id)
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
	err := db.QueryRow(sql, id).Scan(&t.roof, &t.AppID, &t.Author, &t.Prompt, &t.ImgId, &t.ImgPath, &t.Done)
	if err != nil {
		return err
	}
	log.Printf("ticket #%d loaded", id)
	return nil
}

func loadTicket(sn string, id int) (t *Ticket, err error) {
	t = newTicket(sn, AppID(0))
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
	return getDb(t.roof)
}
