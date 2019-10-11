package storage

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-imsto/imagid"
	"github.com/go-imsto/imsto/config"
	iimg "github.com/go-imsto/imsto/image"
	"github.com/go-imsto/imsto/storage/imagio"
	cdb "github.com/go-imsto/imsto/storage/types"
)

const (
	// ViewName ...
	ViewName = "show"
)

var (
	ErrWriteFailed = errors.New("Err: Write file failed")
	ErrEmptyRoof   = errors.New("empty roof")
	ErrEmptyID     = errors.New("empty id")
	ErrZeroSize    = errors.New("zero size")
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
	p        *imagio.Param
	roof     string
	src      string
	dst      string
	id       imagid.IID
	isOrig   bool
	lock     FLock
	name     string
	length   int64
	modified time.Time
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
		return fmt.Errorf("Fatal error: open %s failed", o.p.Path)
	}
	defer file.Close()
	c(file)
	return nil
}

func (o *outItem) prepare() (err error) {
	o.Lock()
	defer o.Unlock()
	orgFile := o.srcName()

	if fi, fe := os.Stat(o.dst); fe == nil && fi.Size() > 0 && o.p.Mop == "" {
		o.length = fi.Size()
		o.modified = fi.ModTime()
		return
	}

	if fi, fe := os.Stat(orgFile); fe != nil && os.IsNotExist(fe) || fe == nil && fi.Size() == 0 {
		mw := NewMetaWrapper(o.roof)
		var entry *mapItem
		entry, err = mw.GetMapping(o.id.String())
		if err != nil {
			// log.Print(err)
			err = NewHttpError(404, err.Error())
			return
		}
		o.roof = entry.roof()
		// log.Printf("got %s", entry.Path)
		roof := o.roof
		// thumb_path := config.GetValue(roof, "thumb_path")
		if o.src != storedPath(entry.Path) { // 302 found
			newPath := path.Join(ViewName, o.p.SizeOp, entry.Path)
			ie := NewHttpError(302, "Found "+newPath)
			ie.Path = newPath
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

		err = Dump(entry.Path, roof, orgFile)
		if err != nil {
			log.Printf("dump fail: %s", err)
			return NewHttpError(404, err.Error())
		}
		if fi, fe := os.Stat(orgFile); fe != nil {
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

	if o.p.Mop != "" {
		if o.p.Mop == "w" {
			orgFile := path.Join(o.thumbRoot(), o.p.SizeOp, o.src)
			dstFile := path.Join(o.thumbRoot(), o.p.SizeOp+"w", o.src)
			watermarkFile := path.Join(config.Root(), config.GetValue(o.roof, "watermark"))
			copyright := config.GetValue(o.roof, "copyright")
			opacity := config.GetInt(o.roof, "watermark_opacity")
			waterOption := iimg.WaterOption{
				Pos:      iimg.Golden,
				Filename: watermarkFile,
				Opacity:  iimg.Opacity(opacity),
			}
			if copyright != "" {
				waterOption.Copyright = path.Join(config.Root(), copyright)
			}
			err = iimg.WatermarkFile(orgFile, dstFile, waterOption)
			if err != nil {
				log.Printf("watermark error: %s", err)
			}
			o.dst = dstFile
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
	// mode := o.m["size"][0:1]
	dimension := o.p.SizeOp[1:]
	// log.Printf("mode %s, dimension %s", mode, dimension)
	supportSize := strings.Split(config.GetValue(o.roof, "support_size"), ",")
	if !stringInSlice(dimension, supportSize) {
		err = NewHttpError(400, fmt.Sprintf("Unsupported size: %s", dimension))
		return
	}

	var topt = iimg.ThumbOption{Width: o.p.Width, Height: o.p.Height, IsFit: true}
	if o.p.Mode == "c" {
		topt.IsCrop = true
	} else if o.p.Mode == "w" {
		topt.MaxWidth = o.p.Width
	} else if o.p.Mode == "h" {
		topt.MaxHeight = o.p.Height
	}
	logger().Infow("thumbnail starting", "roof", o.roof, "name", o.name, "opt", topt)
	err = iimg.ThumbnailFile(o.srcName(), o.dst, topt)
	if err != nil {
		logger().Infow("iimg.ThumbnailFile fail", "src", o.src, "name", o.name, "opt", topt, "err", err)
		return
	}

	if o.p.Mop == "w" && o.p.Width < 100 {
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

// LoadPath ...
func LoadPath(u string) (item *outItem, err error) {
	// log.Printf("load: %s", url)
	var p *imagio.Param
	p, err = imagio.ParseFromPath(u)
	if err != nil {
		logger().Infow("bad url", "url", u, "err", err)
		return
	}
	logger().Debugw("parsed", "param", p)
	item = &outItem{
		p:      p,
		src:    p.Path,
		isOrig: p.IsOrig,
	}

	orgFile := item.srcName()
	if item.isOrig {
		item.dst = orgFile
	} else {
		dstPath := fmt.Sprintf("%s/%s", p.SizeOp, item.src)
		item.dst = path.Join(item.thumbRoot(), dstPath)
	}

	dir := path.Dir(orgFile)
	err = os.MkdirAll(dir, os.FileMode(0755))

	item.lock, err = NewFLock(orgFile + ".lock")
	if err != nil {
		log.Printf("create lock error: %s", err)
		return
	}
	err = item.prepare()
	if err != nil {
		logger().Warnw("prepare fail", "item", item, "err", err)
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

func Dump(key, roof, file string) error {
	logger().Infow("pulling", "roof", roof, "path", key)

	data, err := PullBlob(key, roof)
	if err != nil {
		logger().Warnw("pull fail", "roof", roof, "err", err)
		return err
	}
	logger().Infow("pulled", "roof", roof, "bytes", len(data))
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
	data, err = ioutil.ReadFile(entry.origFullname())
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

func PrepareReader(r io.ReadSeeker, name string, modified uint64) (entry *Entry, err error) {

	entry, err = NewEntryReader(r, name)
	if err != nil {
		return
	}
	entry.Modified = modified
	return
}

func StoredReader(r io.ReadSeeker, name, roof string, modified uint64) (entry *Entry, err error) {
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

func ParseTags(s string) (cdb.StringArray, error) {
	return strings.Split(strings.ToLower(s), ","), nil
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

// deprecated
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
