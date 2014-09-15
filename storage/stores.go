package storage

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	// "net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
	"wpst.me/calf/config"
	cdb "wpst.me/calf/db"
	iimg "wpst.me/calf/image"
)

const (
	image_url_regex = `(?P<tp>[a-z_][a-z0-9_-]*)/(?P<size>[scwh]\d{2,4}(?P<x>x\d{2,4})?|orig)(?P<mop>[a-z])?/(?P<t1>[a-z0-9]{2})/?(?P<t2>[a-z0-9]{2})/?(?P<t3>[a-z0-9]{8,36})\.(?P<ext>gif|jpg|jpeg|png)$`
)

var (
	ire            = regexp.MustCompile(image_url_regex)
	ErrWriteFailed = errors.New("Err: Write file failed")
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

// var lockes = make(map[string]sync.Locker)

// TODO:: clean lockes delay

// temporary item for http read
type outItem struct {
	m        harg
	roof     string
	src      string
	dst      string
	id       *EntryId
	isOrig   bool
	lock     FLock
	name     string
	length   int64
	modified time.Time
}

func newOutItem(url string) (oi *outItem, err error) {

	var m harg
	m, err = parsePath(url)
	if err != nil {
		log.Print(err)
		return
	}
	// log.Print(m)
	roof := getThumbRoof(m["tp"])
	// log.Printf("check roof from %s: %s", m["tp"], roof)
	var id *EntryId
	id, err = NewEntryId(m["t1"] + m["t2"] + m["t3"])
	if err != nil {
		log.Printf("invalid id: %s", err)
		return
	}
	// log.Printf("id: %s", id)
	// thumb_root := config.GetValue(roof, "thumb_root")

	src := fmt.Sprintf("%s/%s/%s.%s", m["t1"], m["t2"], m["t3"], m["ext"])
	// src := path.Join(thumb_root, "orig", org_path)
	isOrig := m["size"] == "orig"

	oi = &outItem{
		m:      m,
		id:     id,
		roof:   roof,
		src:    src,
		isOrig: isOrig,
		name:   fmt.Sprintf("%s%s", id, m["ext"]),
	}

	org_file := oi.srcName()
	if isOrig {
		oi.dst = org_file
	} else {
		dst_path := fmt.Sprintf("%s/%s", m["size"], oi.src)
		oi.dst = path.Join(oi.thumbRoot(), dst_path)
	}

	dir := path.Dir(org_file)
	err = os.MkdirAll(dir, os.FileMode(0755))

	oi.lock, err = NewFLock(org_file + ".lock")
	if err != nil {
		log.Printf("create lock error: %s", err)
		return
	}

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

func (o *outItem) Lock() error {
	return o.lock.Lock()
	// return syscall.Flock(int(o.dst_f.Fd()), syscall.LOCK_EX)
}

func (o *outItem) Unlock() error {
	return o.lock.Unlock()
	// return syscall.Flock(int(o.dst_f.Fd()), syscall.LOCK_UN)
}

func (o *outItem) Walk(c func(file io.ReadSeeker)) error {
	file, err := os.Open(o.dst)
	if err != nil {
		return err
	}
	if file == nil {
		return fmt.Errorf("Fatal error: open %s failed", o.name)
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

	if fi, fe := os.Stat(o.dst); fe == nil && fi.Size() > 0 && o.m["mop"] == "" {
		o.length = fi.Size()
		o.modified = fi.ModTime()
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
		// log.Printf("got %s", entry.Path)
		roof := o.roof
		thumb_path := config.GetValue(roof, "thumb_path")
		if o.src != entry.storedPath() { // 302 found
			new_path := path.Join(thumb_path, o.m["size"], entry.Path)
			ie := NewHttpError(302, "Found "+new_path)
			ie.Path = new_path
			err = ie
			return
		}

		if len(entry.Roofs) > 0 {
			roof0 := fmt.Sprint(entry.Roofs[0])
			if roof != roof0 {
				log.Printf("mismatch roof: %s => %s", o.roof, roof0)
				roof = roof0
			}
		}

		err = Dump(entry, roof, org_file)
		if err != nil {
			log.Printf("dump fail: ", err)
			return NewHttpError(404, err.Error())
		}
		if fi, fe := os.Stat(org_file); fe != nil {
			if os.IsNotExist(fe) || fi.Size() == 0 {
				err = ErrWriteFailed
				return
			}
		}
	}

	err = o.thumbnail()
	if err != nil {
		return
	}

	if o.m["mop"] != "" {
		if o.m["mop"] == "w" {
			org_file := path.Join(o.thumbRoot(), o.m["size"], o.src)
			dst_file := path.Join(o.thumbRoot(), o.m["size"]+"w", o.src)
			watermark_file := path.Join(config.Root(), config.GetValue(o.roof, "watermark"))
			copyright := config.GetValue(o.roof, "copyright")
			opacity := config.GetInt(o.roof, "watermark_opacity")
			waterOption := iimg.WaterOption{
				Pos:      iimg.Golden,
				Filename: watermark_file,
				Opacity:  iimg.Opacity(opacity),
			}
			if copyright != "" {
				waterOption.Copyright = path.Join(config.Root(), copyright)
			}
			err = iimg.WatermarkFile(org_file, dst_file, waterOption)
			if err != nil {
				log.Printf("watermark error: %s", err)
			}
			o.dst = dst_file
		}
	}
	if fi, fe := os.Stat(o.dst); fe == nil && fi.Size() > 0 {
		o.length = fi.Size()
		o.modified = fi.ModTime()
		return
	}
	o.modified = time.Now()
	return
}

func (o *outItem) thumbnail() (err error) {
	if o.isOrig {
		return
	}

	if fi, fe := os.Stat(o.dst); fe == nil && fi.Size() > 0 {
		// log.Print("thumbnail already done")
		return
	}
	mode := o.m["size"][0:1]
	dimension := o.m["size"][1:]
	// log.Printf("mode %s, dimension %s", mode, dimension)
	support_size := strings.Split(config.GetValue(o.roof, "support_size"), ",")
	if !stringInSlice(dimension, support_size) {
		err = NewHttpError(400, fmt.Sprintf("Unsupported size: %s", dimension))
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
	log.Printf("[%s] thumbnail(%s %s) starting", o.roof, o.name, topt)
	err = iimg.ThumbnailFile(o.srcName(), o.dst, topt)
	if err != nil {
		log.Printf("iimg.ThumbnailFile(%s,%s,%s) error: %s", o.src, o.Name, topt, err)
		return
	}

	if o.m["mop"] == "w" && width < 100 {
		return NewHttpError(400, "bad size with watermark")
	}
	return
}

func (o *outItem) Name() string {
	// return o.dst
	return o.name
}

func (o *outItem) Size() int64 {
	return o.length
}

func (o *outItem) Modified() time.Time {
	return o.modified
}

type harg map[string]string

func parsePath(s string) (m harg, err error) {
	if !ire.MatchString(s) {
		err = NewHttpError(400, fmt.Sprintf("Invalid Path: %s", s))
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
		log.Printf("prepare error: %s", err)
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

func Dump(e *Entry, roof, file string) error {
	log.Printf("[%s] pulling: '%s'", roof, e.Path)

	data, err := PullBlob(e, roof)
	if err != nil {
		return err
	}
	log.Printf("[%s] pulled: %d bytes", roof, len(data))
	return SaveFile(file, data)
}

func ReadyDir(filename string) error {
	dir := path.Dir(filename)
	return os.MkdirAll(dir, os.FileMode(0755))
}

func SaveFile(filename string, data []byte) error {
	if err := ReadyDir(filename); err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, os.FileMode(0644))
}

func PopReadyDone() (entry *Entry, err error) {
	entry, err = popPrepared()
	if err != nil {
		return
	}
	log.Printf("poped %s", entry.Path)

	var data []byte
	data, err = ioutil.ReadFile(entry.origName())
	if err != nil {
		return
	}
	err = entry.fill(data)
	if err != nil {
		return
	}
	err = entry._save(entry.roof())
	return
}

func PrepareReader(r io.Reader, name string, modified uint64) (entry *Entry, err error) {
	var data []byte
	data, err = ioutil.ReadAll(r)
	if err != nil {
		return
	}
	entry, err = NewEntry(data, name)

	if err != nil {
		return
	}
	entry.Modified = modified
	return
}

func StoredReader(r io.Reader, name, roof string, modified uint64) (entry *Entry, err error) {
	entry, err = PrepareReader(r, name, modified)
	if err != nil {
		return
	}
	err = entry.Store(roof)
	return
}

func PrepareFile(file, name string) (entry *Entry, err error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return nil, fmt.Errorf("invalid: '%s' is a dir", file)
	}

	if name == "" {
		name = path.Base(file)
	}

	modified := uint64(fi.ModTime().Unix())

	return PrepareReader(f, name, modified)
}

func StoredFile(file, name, roof string) (entry *Entry, err error) {
	entry, err = PrepareFile(file, name)
	if err != nil {
		return
	}
	err = entry.Store(roof)
	return
}

func ParseTags(s string) (cdb.Qarray, error) {
	return cdb.NewQarrayText(s)
	// qtags, err := cdb.NewQarrayText(s)
	// if err == nil {
	// 	return qtags.ToStringSlice(), nil
	// }
	// return nil, err
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

var thumbRoofs = make(map[string]string)

func loadThumbRoofs() error {
	for sec, _ := range config.Sections() {
		s := config.GetValue(sec, "thumb_path")
		tp := strings.TrimPrefix(s, "/")
		if _, ok := thumbRoofs[tp]; !ok {
			thumbRoofs[tp] = sec
		} else {
			return fmt.Errorf("duplicate 'thumb_path=%s' in config", s)
			// log.Printf("duplicate thumb_root in config")
		}
	}
	return nil
}

func getThumbRoof(s string) string {
	tp := strings.Trim(s, "/")
	if v, ok := thumbRoofs[tp]; ok {
		return v
	}
	return ""
}

func init() {
	config.AtLoaded(loadThumbRoofs)
}
