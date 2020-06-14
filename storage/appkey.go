package storage

import (
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	base "github.com/liut/baseconv"
)

const (
	saltSize = 32
)

type App struct {
	Id       AppID     `json:"id,omitempty"`
	Version  VerID     `json:"api_ver,omitempty"`
	ApiKey   string    `json:"api_key,omitempty"`
	ApiSalt  string    `json:"api_salt,omitempty"`
	Name     string    `json:"name,omitempty"`
	Created  time.Time `json:"created,omitempty"`
	Disabled bool      `json:"disabled,omitempty"`
}

func NewApp(name string) (app *App) {
	app = &App{Name: name}
	t := time.Now()
	h := md5.New()
	io.WriteString(h, name)
	io.WriteString(h, fmt.Sprintf("%x", t.UnixNano()))
	s := fmt.Sprintf("%x", h.Sum(nil))
	api_key, _ := base.Convert(s, 16, 62)
	log.Printf("new api_key '%s' for %s", api_key, name)
	app.ApiKey = api_key
	salt, err := newSalt()
	if err != nil {
		log.Print(err)
		return
	}
	app.ApiSalt = salt
	return
}

func LoadApp(api_key string) (app *App, err error) {
	if api_key == "" {
		err = fmt.Errorf("api_key is empty")
		return
	}
	app = &App{ApiKey: api_key}
	err = app.load()
	return
}

func (a *App) load() error {

	db := getDb()

	sql := "SELECT id, name, api_ver, api_salt, disabled, created FROM apps WHERE api_key = $1 LIMIT 1"
	rows, err := db.Query(sql, a.ApiKey)
	if err != nil {
		return err
	}

	if !rows.Next() {
		return fmt.Errorf("api_key not found or disabled")
	}

	return rows.Scan(&a.Id, &a.Name, &a.Version, &a.ApiSalt, &a.Disabled, &a.Created)

}

func (a *App) Save() error {
	qs := func(tx *sql.Tx) (err error) {
		sql := "SELECT app_save($1, $2, $3, $4);"
		var ret int
		err = tx.QueryRow(sql, a.Name, a.ApiKey, a.ApiSalt, a.Version).Scan(&ret)
		if err == nil {
			log.Printf("app saved: %v\n", ret)
			if ret > 0 {
				a.Id = AppID(ret)
			}
		}
		return
	}
	return withTxQuery(qs)
}

func (a *App) genToken() (*apiToken, error) {
	salt, err := a.saltBytes()
	if err != nil {
		return nil, err
	}

	return newToken(a.Version, a.Id, salt)
}

func (a *App) RequestNewToken(uid int) string {
	t, err := a.genToken()
	if err != nil {
		logger().Warnw("genToken fail", "err", err)
		return ""
	}
	var b = make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(uid))
	t.SetValue(b, VC_TOKEN)
	return t.String()
}

// VerifyToken ...
func (a *App) VerifyToken(token string) (at *apiToken, err error) {
	if token == "" {
		err = errors.New("api: need token argument")
		return
	}

	if at, err = a.genToken(); err != nil {
		err = fmt.Errorf("genToken: %s", err)
		return
	}
	err = at.VerifyString(token)
	return
}

func newSalt() (rs string, err error) {
	rb := make([]byte, saltSize)
	_, err = rand.Read(rb)

	if err != nil {
		return
	}

	rs = base64.URLEncoding.EncodeToString(rb)
	return
}

func (a *App) saltBytes() ([]byte, error) {
	return base64.URLEncoding.DecodeString(a.ApiSalt)
}
