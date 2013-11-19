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
	"os"
	"strconv"
)

type Ticket struct {
	section string
	dsn     string
	table   string
	app     AppId
	author  Author
	prompt  string
	id      int
}

func newTicket(section string, app AppId) *Ticket {
	dsn := config.GetValue(section, "meta_dsn")
	table := config.GetValue(section, "ticket_table")
	// log.Printf("table: %s", table)
	t := &Ticket{section: section, dsn: dsn, table: table, app: app}

	return t
}

func NewTicketRequest(r *http.Request) (t *Ticket, err error) {
	if err = r.ParseForm(); err != nil {
		log.Print("form parse error:", err)
		return
	}

	section := r.PostFormValue("section")
	app := r.PostFormValue("appid")
	var appid uint64
	appid, err = strconv.ParseUint(app, 10, 16)
	if err != nil {
		log.Printf("arg appid error: %s", err)
		appid = 0
	}

	var salt []byte
	salt, err = getApiSalt(section, appid)
	token := NewToken(salt)
	var ok bool
	ok, err = token.VerifyString(r.PostFormValue("token"))
	if err != nil {
		return
	}
	if !ok {
		err = errors.New("Invalid Token")
	}
	t = newTicket(section, AppId(appid))

	user := r.PostFormValue("user")
	var uid uint64
	uid, err = strconv.ParseUint(user, 10, 16)
	if err != nil {
		log.Printf("arg author error: %s", err)
		uid = 0
	}
	t.author = Author(uid)
	t.prompt = r.PostFormValue("prompt")

	return
}

func (t *Ticket) saveNew() error {
	db := t.getDb()
	defer db.Close()

	var id int
	sql := "INSERT INTO " + t.table + "(app_id, author, prompt) VALUES($1, $2, $3) RETURNING id"
	err := db.QueryRow(sql, t.app, t.author, t.prompt).Scan(&id)
	if err != nil {
		return err
	}
	t.id = id
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

func getApiSalt(section string, appid AppId) (salt []byte, err error) {
	k := fmt.Sprintf("IMSTO_API_%d_SALT", appid)
	str := config.GetValue(section, k)
	if str == "" {
		str = os.Getenv(k)
	}

	if str == "" {
		err = fmt.Errorf("%s not found in environment or config", k)
		return
	}

	salt = []byte(str)
	return
}
