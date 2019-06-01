package storage

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"

	"github.com/go-imsto/imsto/config"
)

const (
	defaultMaxMemory = 12 << 20 // 8 MB
	ApiKeyHeader     = "X-Access-Key"
)

func StoredRequest(r *http.Request) (entries map[string][]entryStored, err error) {
	if err = r.ParseMultipartForm(defaultMaxMemory); err != nil {
		log.Print("multipart form parse error:", err)
		return
	}
	cr, e := parseRequest(r, true)
	if e != nil {
		err = e
		return
	}

	lastModified, _ := strconv.ParseUint(r.FormValue("file_ts"), 10, 64)
	log.Printf("lastModified: %v", lastModified)

	form := r.MultipartForm
	if form == nil {
		err = errors.New("browser error: no file select")
		return
	}
	defer form.RemoveAll()

	if len(form.File) == 0 {
		err = errors.New("browser error: input file not found or invalid POST")
		return
	}

	tags, err := ParseTags(r.FormValue("tags"))
	if err != nil {
		log.Print(err)
	}

	entries = make(map[string][]entryStored)
	for k, fhs := range form.File {
		entries[k] = make([]entryStored, len(fhs))
		for i, fh := range fhs {
			log.Printf("%d name: %s, ctype: %s", i, fh.Filename, fh.Header.Get("Content-Type"))
			mime := fh.Header.Get("Content-Type")
			file, fe := fh.Open()
			if fe != nil {
				entries[k][i].Err = fe.Error()
			}

			log.Printf("post %s (%s) size %d\n", fh.Filename, mime, len(fh.Header))
			// entry, ee := NewEntry(data, fh.Filename)
			entry, ee := PrepareReader(file, fh.Filename, lastModified)
			if ee != nil {
				entries[k][i].Err = ee.Error()
				continue
			}
			entry.AppId = cr.app.Id
			entry.Author = cr.author
			// entry.Modified = lastModified
			entry.Tags = tags
			ee = entry.Store(cr.roof)
			if ee != nil {
				log.Printf("%02d stored error: %s", i, ee)
				entries[k][i].Err = ee.Error()
				continue
			}
			log.Printf("%02d [%s]stored %s %s", i, cr.roof, entry.Id, entry.Path)
			if ee == nil && i == 0 && cr.token.vc == VC_TICKET {
				tid := cr.token.GetValuleInt()
				log.Printf("token value: %d", tid)
				var ticket *Ticket
				ticket, ee = loadTicket(cr.roof, int(tid))
				if ee != nil {
					log.Printf("ticket load error: %s", ee)
				}
				ee = ticket.bindEntry(entry)
				if ee != nil {
					log.Printf("ticket bind error: %v", ee)
				}
			}
			entries[k][i] = newStoredEntry(entry, ee)
		}
	}

	return
}

func DeleteRequest(r *http.Request) error {
	dir, id := path.Split(r.URL.Path)
	roof := path.Base(dir)
	if id != "" && roof != "" {
		mw := NewMetaWrapper(roof)
		eid, err := NewEntryId(id)
		if err != nil {
			return err
		}
		err = mw.Delete(*eid)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("invalid url")
}

type custReq struct {
	roof   string
	app    *App
	author Author
	token  *apiToken
}

func parseRequest(r *http.Request, needToken bool) (cr custReq, err error) {
	if r.Form == nil {
		if err = r.ParseForm(); err != nil {
			log.Print("form parse error:", err)
			return
		}
	}

	var (
		roof   string
		author Author
		str    string
		app    *App
		uid    uint64
	)
	roof = r.FormValue("roof")
	if roof == "" {
		log.Print("Waring: parseRequest roof is empty")
	}

	if !config.HasSection(roof) {
		err = fmt.Errorf("roof '%s' not found", roof)
		return
	}

	str = r.Header.Get(ApiKeyHeader)
	if str == "" {
		str = r.FormValue("api_key")
		if str == "" {
			log.Print("Waring: parseRequest api_key is empty")
		}
	}

	app, err = LoadApp(str)
	if err != nil {
		err = fmt.Errorf("arg 'api_key=%s' is invalid: %s", str, err.Error())
		return
	}

	if app.Disabled {
		err = fmt.Errorf("the api_key %s is invalid", str)
		return
	}

	str = r.FormValue("user")
	if str != "" {
		uid, err = strconv.ParseUint(str, 10, 16)
		if err != nil {
			// log.Printf("arg user error: %s", err)
			err = fmt.Errorf("arg 'user=%s' is invalid: %s", str, err.Error())
			return
		}
	}

	author = Author(uid)

	var token *apiToken
	if needToken {
		if token, err = app.genToken(); err != nil {
			err = fmt.Errorf("genToken: %s", err)
			return
		}
		token_str := r.FormValue("token")
		if token_str == "" {
			err = errors.New("api: need token argument")
			return
		}

		if _, err = token.VerifyString(token_str); err != nil {
			return
		}
	}

	cr = custReq{roof, app, author, token}
	return
}
