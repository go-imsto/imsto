package storage

import (
	"calf/config"
	"calf/db"
	iimg "calf/image"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const (
	image_url_regex  = `(?P<size>[scwh]\d{2,4}(?P<x>x\d{2,4})?|orig)(?P<mop>[a-z])?/(?P<t1>[a-z0-9]{2})/(?P<t2>[a-z0-9]{2})/(?P<t3>[a-z0-9]{19,36})\.(?P<ext>gif|jpg|jpeg|png)$`
	defaultMaxMemory = 16 << 20 // 16 MB
)

var (
	ire              = regexp.MustCompile(image_url_regex)
	ErrInvalidUrl    = errors.New("Err: Invalid Url")
	ErrWriteFailed   = errors.New("Err: Write file failed")
	ErrUnsupportSize = errors.New("Err: Unsupported size")
)

type ErrHttpFound struct {
	Path string
}

func (ie ErrHttpFound) Error() string {
	return "Found " + ie.Path
}

type outItem struct {
	DestPath, DestFile, Mime string
	Size                     int
}

type harg map[string]string

func parsePath(s string) (m harg, err error) {
	if !ire.MatchString(s) {
		err = ErrInvalidUrl
		return
	}
	match := ire.FindStringSubmatch(s)
	names := ire.SubexpNames()
	m = make(harg)
	for i, n := range names {
		if n != "" {
			m[n] = match[i]
		}
	}
	return
}

func LoadPath(url, section string) (item outItem, err error) {
	var m harg
	m, err = parsePath(url)
	if err != nil {
		log.Print(err)
		return
	}
	log.Print(m)
	var id *EntryId
	id, err = NewEntryId(m["t1"] + m["t2"] + m["t3"])
	if err != nil {
		log.Print(err)
		return
	}
	log.Printf("id: %s", id)
	thumb_root := config.GetValue(section, "thumb_root")

	org_path := fmt.Sprintf("%s/%s/%s.%s", m["t1"], m["t2"], m["t3"], m["ext"])
	org_file := path.Join(thumb_root, "orig", org_path)

	var fi os.FileInfo
	if fi, err = os.Stat(org_file); err != nil {
		if os.IsNotExist(err) || fi.Size() == 0 {

			mw := NewMetaWrapper(section)
			var e *emap
			e, err = mw.GetMap(*id)
			if err != nil {
				log.Print(err)
				return
			}
			log.Print(e)
			if org_path != e.path { // 302 found
				thumb_path := config.GetValue(section, "thumb_path")
				new_path := path.Join(thumb_path, m["size"], e.path)
				err = ErrHttpFound{Path: new_path}
				return
			}
			log.Printf("fetching file: %s", e.path)

			var em Wagoner
			em, err = FarmEngine(section)
			if err != nil {
				log.Println(err)
				return
			}
			var data []byte
			data, err = em.Get(e.path)
			log.Printf("fetched: %d", len(data))
			err = saveFile(org_file, data)
			if err != nil {
				log.Print(err)
				return
			}
			if fi, err = os.Stat(org_file); err != nil {
				if os.IsNotExist(err) || fi.Size() == 0 {
					err = ErrWriteFailed
					return
				}
			}
		}
	}
	var (
		dst_path, dst_file string
	)
	if m["size"] == "orig" {
		dst_path = "orig/" + org_path
		dst_file = org_file
	} else {
		dst_path = fmt.Sprintf("%s/%s", m["size"], org_path)
		dst_file = path.Join(thumb_root, dst_path)

		mode := m["size"][0:1]
		dimension := m["size"][1:]
		log.Printf("mode %s, dimension %s", mode, dimension)
		support_size := strings.Split(config.GetValue(section, "support_size"), ",")
		if !stringInSlice(dimension, support_size) {
			err = ErrUnsupportSize
			return
		}
		var width, height uint
		if m["x"] == "" {
			var d uint64
			d, _ = strconv.ParseUint(dimension, 10, 32)
			width = uint(d)
			height = uint(d)
		} else {
			a := strings.Split(dimension, "x")
			var dw, dh uint64
			dw, _ = strconv.ParseUint(a[0], 10, 32)
			dh, _ = strconv.ParseUint(a[1], 10, 32)
			width = uint(dw)
			height = uint(dh)
		}

		var topt = iimg.ThumbOption{Width: width, Height: height, IsFit: true}
		if mode == "c" {
			topt.IsCrop = true
		} else if mode == "w" {
			topt.MaxWidth = width
		} else if mode == "h" {
			topt.MaxHeight = height
		}
		err = iimg.ThumbnailFile(org_file, dst_file, topt)
		if err != nil {
			log.Print(err)
			return
		}

		// if fi, _ := os.Stat(dst_file); fi.Size() == 0 {
		// 	log.Print("thumbnail dst_file fail")
		// }
	}
	log.Printf("dst_path: %s, dst_file: %s", dst_path, dst_file)

	item = outItem{}
	item.DestPath = dst_path
	item.DestFile = dst_file

	return
}

func stringInSlice(s string, a []string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}

func saveFile(filename string, data []byte) (err error) {
	dir := path.Dir(filename)
	err = os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		return
	}
	err = ioutil.WriteFile(filename, data, os.FileMode(0644))
	return
}

func StoredFile(filename string, section string) (entry *Entry, err error) {
	var fi os.FileInfo
	if fi, err = os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			log.Println(err)
			return
		}
	}

	var data []byte

	data, err = ioutil.ReadFile(filename)

	if err != nil {
		log.Println(err)
		return
	}

	entry, err = newEntry(data, path.Base(filename))

	if err != nil {
		return
	}
	entry.Modified = uint64(fi.ModTime().Unix())
	// err = entry.trek(section)

	err = store(entry, section)
	if err != nil {
		// log.Println(err)
		return
	}

	return
}

type entryStored struct {
	*Entry
	Err string `json:"error,omitempty"`
}

func newStoredEntry(entry *Entry, err error) entryStored {
	var es string
	if err != nil {
		es = err.Error()
	} else {
		es = ""
	}
	return entryStored{entry, es}
}

func StoredRequest(r *http.Request) (entries []entryStored, err error) {
	if err = r.ParseMultipartForm(defaultMaxMemory); err != nil {
		log.Print("form parse error:", err)
		return
	}

	// log.Printf("form: %s", r.Form)
	// log.Printf("postform: %s", r.PostForm)

	var (
		section      string
		appid        AppId
		author       Author
		lastModified uint64
	)

	section, appid, author, err = parseRequest(r)
	if err != nil {
		// log.Print("request error:", err)
		return
	}

	if !config.HasSection(section) {
		err = fmt.Errorf("section '%s' not found", section)
		return
	}

	var token *apiToken
	token, err = getApiToken(section, appid)
	if err != nil {
		return
	}
	token_str := r.FormValue("token")
	if token_str == "" {
		err = errors.New("api: need token argument")
		return
	}
	var ok bool
	ok, err = token.VerifyString(token_str)
	if err != nil {
		return
	}
	if !ok {
		err = errors.New("api: Invalid Token")
		return
	}

	lastModified, _ = strconv.ParseUint(r.FormValue("ts"), 10, 64)
	log.Printf("lastModified: %s", lastModified)

	form := r.MultipartForm
	if form == nil {
		err = errors.New("browser error: no file select")
		return
	}
	defer form.RemoveAll()

	if _, ok := form.File["file"]; !ok {
		err = errors.New("browser error: input 'file' not found")
		return
	}

	n := len(form.File["file"])

	log.Printf("%d files", n)

	entries = make([]entryStored, n)
	for i, fh := range form.File["file"] {
		log.Printf("%d name: %s, ctype: %s", i, fh.Filename, fh.Header.Get("Content-Type"))
		mime := fh.Header.Get("Content-Type")
		file, fe := fh.Open()
		if fe != nil {
			entries[i].Err = fe.Error()
		}

		data, ee := ioutil.ReadAll(file)
		if ee != nil {
			entries[i].Err = ee.Error()
			continue
		}
		log.Printf("post %s (%s) size %d\n", fh.Filename, mime, len(data))
		entry, ee := newEntry(data, fh.Filename)
		if ee != nil {
			entries[i].Err = ee.Error()
			continue
		}
		entry.AppId = appid
		entry.Author = author
		entry.Modified = lastModified
		ee = store(entry, section)
		if ee != nil {
			log.Printf("%02d stored error: %s", i, ee)
		}
		if ee == nil && i == 0 && token.vc == VC_TICKET {
			// TODO: upate ticket
			// ticket := newTicket(section, appid)
			tid := token.GetValuleInt()
			log.Printf("token value: %d", tid)
			var ticket *Ticket
			ticket, ee = loadTicket(section, int(tid))
			if ee != nil {
				log.Printf("ticket load error: ", ee)
			}
			ee = ticket.bindEntry(entry)
			if ee != nil {
				log.Printf("ticket bind error: %v", ee)
			}
		}
		entries[i] = newStoredEntry(entry, ee)
	}

	return
}

func store(e *Entry, section string) (err error) {
	err = e.Trek(section)
	if err != nil {
		// log.Println(err)
		return
	}
	log.Printf("new id: %v, size: %d, path: %v\n", e.Id, e.Size, e.Path)

	data := e.Blob()
	size := len(data)
	log.Printf("blob length: %d", size)

	var em Wagoner
	em, err = FarmEngine(section)
	if err != nil {
		log.Println(err)
		return
	}

	var sev db.Hstore
	sev, err = em.Put(e.Path, data, e.Meta.Hstore())
	if err != nil {
		log.Println(err)
		return
	}

	e.sev = sev

	mw := NewMetaWrapper(section)
	err = mw.Store(e)
	// fmt.Println("mw", mw)
	if err != nil {
		log.Println(err)
		return
	}
	return
}

func DeleteRequest(r *http.Request) error {
	dir, id := path.Split(r.URL.Path)
	section := path.Base(dir)
	if id != "" && section != "" {
		mw := NewMetaWrapper(section)
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

func parseRequest(r *http.Request) (section string, appid AppId, author Author, err error) {
	if r.Form == nil {
		if err = r.ParseForm(); err != nil {
			log.Print("form parse error:", err)
			return
		}
	}
	var (
		str string
		aid uint64
		uid uint64
	)
	section = r.FormValue("section")
	str = r.FormValue("app")
	if str != "" {
		aid, err = strconv.ParseUint(str, 10, 16)
		if err != nil {
			// log.Printf("arg app error: %s", err)
			err = fmt.Errorf("arg 'app=%s' is invalid: %s", str, err.Error())
			return
		}
	}

	appid = AppId(aid)

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
	return
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

func getApiToken(section string, appid AppId) (token *apiToken, err error) {
	var salt []byte
	salt, err = getApiSalt(section, appid)
	if err != nil {
		return
	}
	ver := apiVer(0)
	token, err = newToken(ver, appid, salt)
	return
}
