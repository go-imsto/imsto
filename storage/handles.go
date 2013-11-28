package storage

import (
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
	"sync"
	"wpst.me/calf/config"
	iimg "wpst.me/calf/image"
)

const (
	image_url_regex  = `(?P<tp>[a-z][a-z0-9]*)/(?P<size>[scwh]\d{2,4}(?P<x>x\d{2,4})?|orig)(?P<mop>[a-z])?/(?P<t1>[a-z0-9]{2})/(?P<t2>[a-z0-9]{2})/(?P<t3>[a-z0-9]{19,36})\.(?P<ext>gif|jpg|jpeg|png)$`
	defaultMaxMemory = 16 << 20 // 16 MB
)

var (
	ire              = regexp.MustCompile(image_url_regex)
	ErrInvalidUrl    = errors.New("Err: Invalid Url")
	ErrWriteFailed   = errors.New("Err: Write file failed")
	ErrUnsupportSize = errors.New("Err: Unsupported size")
)

type HttpError struct {
	Code int
	Text string
	Path string
}

func (ie *HttpError) Error() string {
	return fmt.Sprintf("%d: %s", ie.Code, ie.Text)
}

func NewHttpError(code int, text string) *HttpError {
	return &HttpError{Code: code, Text: text}
}

var lockes = make(map[string]sync.Locker)

// TODO:: clean lockes delay

// temporary item for http read
type outItem struct {
	k      string
	m      harg
	roof   string
	src    string
	dst    string
	id     *EntryId
	isOrig bool
}

func newOutItem(url string) (oi *outItem, err error) {

	var m harg
	m, err = parsePath(url)
	if err != nil {
		log.Print(err)
		return
	}
	// log.Print(m)
	roof := config.ThumbRoof(m["tp"])
	var id *EntryId
	id, err = NewEntryId(m["t1"] + m["t2"] + m["t3"])
	if err != nil {
		log.Printf("invalid id: %s", err)
		return
	}
	// log.Printf("id: %s", id)
	// thumb_root := config.GetValue(roof, "thumb_root")

	k := roof + id.String()

	src := fmt.Sprintf("%s/%s/%s.%s", m["t1"], m["t2"], m["t3"], m["ext"])
	// src := path.Join(thumb_root, "orig", org_path)
	oi = &outItem{k: k, m: m, id: id, roof: roof, src: src, isOrig: m["size"] == "orig"}

	return
}

func (o *outItem) cfg(s string) string {
	return config.GetValue(o.roof, s)
}

func (o *outItem) srcName() string {
	return path.Join(o.thumbRoot(), "orig", o.src)
}

func (o *outItem) thumbRoot() string {
	return o.cfg("thumb_root")
}

func (o *outItem) Lock() {
	var (
		lkr sync.Locker
		ok  bool
	)
	if lkr, ok = lockes[o.k]; !ok {
		lkr = new(sync.Mutex)
		lockes[o.k] = lkr
	}
	lkr.Lock()
}

func (o *outItem) Unlock() {
	if lkr, ok := lockes[o.k]; ok {
		lkr.Unlock()
	}
}

func (o *outItem) Walk(c func(file http.File)) error {
	// o.Lock()
	// defer func() {
	// o.Unlock()
	// delete(lockes, o.k)
	// }()
	file, err := os.Open(o.dst)
	if err != nil {
		return err
	}
	if file == nil {
		return fmt.Errorf("Fatal error: open %s failed", o.Name)
	}
	defer file.Close()
	c(file)
	return nil
}

func (o *outItem) prepare() (err error) {
	o.Lock()
	defer o.Unlock()
	org_file := o.srcName()

	if o.isOrig {
		o.dst = org_file
	} else {
		dst_path := fmt.Sprintf("%s/%s", o.m["size"], o.src)
		o.dst = path.Join(o.thumbRoot(), dst_path)
	}

	if fi, fe := os.Stat(o.dst); fe == nil && fi.Size() > 0 {
		return
	}

	if fi, fe := os.Stat(org_file); fe != nil && os.IsNotExist(fe) || fe == nil && fi.Size() == 0 {
		mw := NewMetaWrapper(o.roof)
		var entry *Entry
		entry, err = mw.GetEntry(*o.id)
		if err != nil {
			// log.Print(err)
			err = NewHttpError(404, err.Error())
			return
		}
		// log.Printf("got %s", entry)
		if o.src != entry.Path { // 302 found
			thumb_path := config.GetValue(o.roof, "thumb_path")
			new_path := path.Join(thumb_path, o.m["size"], entry.Path)
			ie := NewHttpError(302, "Found "+new_path)
			ie.Path = new_path
			err = ie
			return
		}
		log.Printf("fetching [%s] file: '%s'", o.roof, entry.Path)

		var em Wagoner
		em, err = FarmEngine(o.roof)
		if err != nil {
			log.Println(err)
			return
		}
		var data []byte
		data, err = em.Get(entry.Path)
		if err != nil {
			err = NewHttpError(404, err.Error())
			return
		}
		log.Printf("fetched: %d", len(data))
		err = saveFile(org_file, data)
		if err != nil {
			log.Printf("save fail: ", err)
			return
		}
		if fi, fe := os.Stat(org_file); fe != nil {
			if os.IsNotExist(fe) || fi.Size() == 0 {
				err = ErrWriteFailed
				return
			}
		}
	}

	err = o.thumbnail()
	return
}

func (o *outItem) thumbnail() (err error) {
	if o.isOrig {
		return
	}
	// o.Lock()
	// defer func() {
	// 	o.Unlock()
	// 	delete(cachedItems, o.k)
	// }()

	if fi, fe := os.Stat(o.dst); fe == nil && fi.Size() > 0 {
		// log.Print("thumbnail already done")
		return
	}
	mode := o.m["size"][0:1]
	dimension := o.m["size"][1:]
	// log.Printf("mode %s, dimension %s", mode, dimension)
	support_size := strings.Split(config.GetValue(o.roof, "support_size"), ",")
	if !stringInSlice(dimension, support_size) {
		err = NewHttpError(400, ErrUnsupportSize.Error())
		return
	}
	var width, height uint
	if o.m["x"] == "" {
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
	log.Printf("outItem.thumbnail(%dx%d %v) starting", topt.Width, topt.Height, topt.IsCrop)
	err = iimg.ThumbnailFile(o.srcName(), o.dst, topt)
	if err != nil {
		log.Printf("iimg.ThumbnailFile(%s,%s,%s) error: %s", o.src, o.Name, topt, err)
		return
	}
	return
}

func (o *outItem) Name() string {
	return o.dst
}

type harg map[string]string

func parsePath(s string) (m harg, err error) {
	if !ire.MatchString(s) {
		err = NewHttpError(400, ErrInvalidUrl.Error())
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

func LoadPath(url string) (item *outItem, err error) {
	// log.Printf("load: %s", url)
	item, err = newOutItem(url)
	if err != nil {
		log.Print(err)
		return
	}
	err = item.prepare()
	if err != nil {
		log.Print(err)
		return
	}
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

func StoredFile(filename string, roof string) (entry *Entry, err error) {
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
	// err = entry.trek(roof)

	err = entry.store(roof)
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
		roof         string
		appid        AppId
		author       Author
		lastModified uint64
	)

	roof, appid, author, err = parseRequest(r)
	if err != nil {
		// log.Print("request error:", err)
		return
	}

	if !config.HasSection(roof) {
		err = fmt.Errorf("roof '%s' not found", roof)
		return
	}

	var token *apiToken
	token, err = getApiToken(roof, appid)
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
		ee = entry.store(roof)
		if ee != nil {
			log.Printf("%02d stored error: %s", i, ee)
			entries[i].Err = ee.Error()
			continue
		}
		log.Printf("stored %s %s", entry.Id, entry.Path)
		if ee == nil && i == 0 && token.vc == VC_TICKET {
			// TODO: upate ticket
			// ticket := newTicket(roof, appid)
			tid := token.GetValuleInt()
			log.Printf("token value: %d", tid)
			var ticket *Ticket
			ticket, ee = loadTicket(roof, int(tid))
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

func parseRequest(r *http.Request) (roof string, appid AppId, author Author, err error) {
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
	roof = r.FormValue("roof")
	if roof == "" {
		log.Print("Waring: parseRequest roof is empty")
	}
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

func getApiSalt(roof string, appid AppId) (salt []byte, err error) {
	k := fmt.Sprintf("IMSTO_API_%d_SALT", appid)
	str := config.GetValue(roof, k)
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

func getApiToken(roof string, appid AppId) (token *apiToken, err error) {
	var salt []byte
	salt, err = getApiSalt(roof, appid)
	if err != nil {
		return
	}
	ver := apiVer(0)
	token, err = newToken(ver, appid, salt)
	return
}
