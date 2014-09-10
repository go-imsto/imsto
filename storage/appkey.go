package storage

import (
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	_ "database/sql/driver"
	"encoding/base64"
	"fmt"
	_ "github.com/lib/pq"
	"io"
	"log"
	"time"
	"wpst.me/calf/base"
	"wpst.me/calf/config"
)

const (
	salt_size = 32
)

type App struct {
	Id       AppId      `json:"id,omitempty"`
	Version  ApiVersion `json:"api_ver,omitempty"`
	ApiKey   string     `json:"api_key,omitempty"`
	ApiSalt  string     `json:"api_salt,omitempty"`
	Name     string     `json:"name,omitempty"`
	Created  time.Time  `json:"created,omitempty"`
	Disabled bool       `json:"disabled,omitempty"`
}

func NewApp(name string) (app *App) {
	app = &App{Name: name}
	t := time.Now()
	h := md5.New()
	io.WriteString(h, name)
	io.WriteString(h, fmt.Sprintf("%x", t.UnixNano()))
	s := fmt.Sprintf("%x", h.Sum(nil))
	api_key, _ := base.BaseConvert(s, 16, 62)
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

func (this *App) load() error {

	db := this.getDb()
	defer db.Close()

	sql := "SELECT id, name, api_ver, api_salt, disabled, created FROM apps WHERE api_key = $1 LIMIT 1"
	rows, err := db.Query(sql, this.ApiKey)
	if err != nil {
		return err
	}

	if !rows.Next() {
		return fmt.Errorf("not found or disabled")
	}

	return rows.Scan(&this.Id, &this.Name, &this.Version, &this.ApiSalt, &this.Disabled, &this.Created)

}

func (this *App) Save() error {
	db := this.getDb()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	sql := "SELECT app_save($1, $2, $3, $4);"

	var ret int
	err = tx.QueryRow(sql, this.Name, this.ApiKey, this.ApiSalt, this.Version).Scan(&ret)

	if err != nil {
		tx.Rollback()
		return err
	}

	log.Printf("app saved: %v\n", ret)

	if ret > 0 {
		this.Id = AppId(ret)
	}

	tx.Commit()
	return nil
}

func (this *App) getDb() *sql.DB {
	dsn := config.GetValue("", "meta_dsn")
	db, err := sql.Open("postgres", dsn)

	if err != nil {
		log.Fatalf("open db error: %s", err)
	}
	return db
}

func (this *App) genToken() (*apiToken, error) {
	salt, err := this.saltBytes()
	if err != nil {
		return nil, err
	}

	return newToken(this.Version, this.Id, salt)
}

func newSalt() (rs string, err error) {
	rb := make([]byte, salt_size)
	_, err = rand.Read(rb)

	if err != nil {
		return
	}

	rs = base64.URLEncoding.EncodeToString(rb)
	return
}

func (this *App) saltBytes() ([]byte, error) {
	return base64.URLEncoding.DecodeString(this.ApiSalt)
}
