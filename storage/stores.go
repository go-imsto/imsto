package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	iimg "github.com/go-imsto/imagi"
	"github.com/go-imsto/imid"
	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/storage/imagio"
	cdb "github.com/go-imsto/imsto/storage/types"
	"github.com/go-imsto/imsto/utils"
)

// consts Cate of Key
const (
	CatView  = "show"
	CatStore = "stores"
	CatThumb = "thumb"
)

// errors
var (
	ErrWriteFailed = errors.New("Err: Write file failed")
	ErrEmptyRoof   = errors.New("empty roof")
	ErrEmptyID     = errors.New("empty id")
	ErrZeroSize    = errors.New("zero size")
	ErrInvalidRoof = errors.New("empty roof")
)

// HttpError ...
type HttpError struct {
	Code int
	Text string
	Path string
}

// Error ...
func (ie *HttpError) Error() string {
	return fmt.Sprintf("%d: %s", ie.Code, ie.Text)
}

// NewHttpError ...
func NewHttpError(code int, text string) *HttpError {
	return &HttpError{Code: code, Text: text}
}

// File ...
type File interface {
	io.Reader
	io.Seeker

	Name() string
	Size() int64
	Modified() time.Time
}

// WalkFunc ..
type WalkFunc func(f File)

// Walker ...
type Walker interface {
	Walk(c WalkFunc) error
}

type file struct {
	dst      string
	name     string
	length   int64
	modified time.Time
	*os.File
}

func (f *file) Name() string {
	return f.name
}

func (f *file) Size() int64 {
	return f.length
}

func (f *file) Modified() time.Time {
	return f.modified
}

// temporary item for http read
type outItem struct {
	p        *imagio.Param
	roof     string
	src      string
	dst      string
	id       imid.IID
	isOrig   bool
	lock     utils.FLock
	name     string
	length   int64
	modified time.Time
	root     string
	origFile string
}

func (o *outItem) Walk(c WalkFunc) error {
	fp, err := os.Open(o.dst)
	if err != nil {
		return err
	}
	if fp == nil {
		return fmt.Errorf("Fatal error: open %s failed", o.p.Path)
	}
	defer fp.Close()
	c(&file{
		File:     fp,
		name:     o.name,
		length:   o.length,
		modified: o.modified,
	})
	return nil
}

func (o *outItem) prepare() (err error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if fi, fe := os.Stat(o.dst); fe == nil && fi.Size() > 0 && o.p.Mop == "" {
		o.length = fi.Size()
		o.modified = fi.ModTime()
		return
	}

	var roof string
	logger().Infow("prepare", "orig", o.origFile)
	if fi, fe := os.Stat(o.origFile); fe != nil && os.IsNotExist(fe) || fe == nil && fi.Size() == 0 {
		logger().Infow("get mapping", "id", o.id)
		mw := NewMetaWrapper(commonRoof)
		var entry *mapItem
		entry, err = mw.GetMapping(o.id.String())
		if err != nil {
			logger().Infow("get mapping fail", "id", o.id, "err", err)
			err = NewHttpError(404, err.Error())
			return
		}
		o.roof = entry.roof()
		roof = o.roof

		if len(entry.Roofs) > 0 {
			roof0 := fmt.Sprint(entry.Roofs[0])
			if roof != roof0 {
				logger().Infow("mismatch roof", "roof", o.roof, "roof0", roof0)
				roof = roof0
			}
		}

		var data []byte
		data, err = entry.pullWith(roof)
		if err != nil {
			return NewHttpError(500, err.Error())
		}
		err = utils.SaveFile(o.origFile, data)
		if err != nil {
			return NewHttpError(500, err.Error())
		}
		if fi, fe := os.Stat(o.origFile); fe != nil {
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
		watermarkFile := config.Current.WatermarkFile
		if o.p.Mop == "w" && watermarkFile != "" {
			orgFile := path.Join(o.root, o.p.SizeOp, o.src)
			dstFile := path.Join(o.root, o.p.SizeOp+"w", o.src)
			waterOption := iimg.WaterOption{
				Pos:      iimg.Golden,
				Filename: watermarkFile,
				Opacity:  iimg.Opacity(config.Current.WatermarkOpacity),
			}
			err = iimg.WatermarkFile(orgFile, dstFile, waterOption)
			if err != nil {
				logger().Infow("watermark fail", "err", err)
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
	supportSize := config.Current.SupportSizes
	if !supportSize.Has(o.p.Width) || !supportSize.Has(o.p.Height) {
		err = NewHttpError(400, fmt.Sprintf("Unsupported size: %s", dimension))
		return
	}

	var topt = iimg.ThumbOption{Width: o.p.Width, Height: o.p.Height, IsFit: true}
	topt.Format = o.p.Ext
	if o.p.Mode == "c" {
		topt.IsCrop = true
	} else if o.p.Mode == "w" {
		topt.MaxWidth = o.p.Width
	} else if o.p.Mode == "h" {
		topt.MaxHeight = o.p.Height
	}
	logger().Infow("thumbnail starting", "roof", o.roof, "name", o.name, "opt", topt)
	err = iimg.ThumbnailFile(o.origFile, o.dst, topt)
	if err != nil {
		logger().Infow("iimg.ThumbnailFile fail", "src", o.src, "name", o.name, "opt", topt, "err", err)
		return
	}

	if o.p.Mop == "w" && o.p.Width < 100 {
		return NewHttpError(400, "bad size with watermark")
	}
	return
}

// storedPath ...
func storedPath(r string) string {
	return imagio.StoredPath(r)
}

// LoadPath ...
func LoadPath(u string) (Walker, error) {
	// log.Printf("load: %s", url)
	p, err := imagio.ParseFromPath(u)
	if err != nil {
		logger().Infow("bad url", "url", u, "err", err)
		return nil, err
	}
	logger().Debugw("parsed", "param", p)
	root := path.Join(config.Current.CacheRoot, CatThumb)
	oi := &outItem{
		p:        p,
		id:       p.ID,
		src:      p.Path,
		isOrig:   p.IsOrig,
		root:     root,
		origFile: path.Join(root, "orig", p.Path),
	}

	if oi.isOrig {
		oi.dst = oi.origFile
	} else {
		dstPath := fmt.Sprintf("%s/%s", p.SizeOp, oi.src)
		oi.dst = path.Join(oi.root, dstPath)
	}

	err = utils.ReadyDir(oi.origFile)
	if err != nil {
		logger().Infow("ready dir fail", "err", err)
		return nil, err
	}

	oi.lock, err = utils.NewFLock(oi.origFile + ".lock")
	if err != nil {
		logger().Infow("create lock fail", "err", err)
		return nil, err
	}
	err = oi.prepare()
	if err != nil {
		logger().Warnw("prepare fail", "param", oi.p, "err", err)
		return nil, err
	}
	return oi, nil
}

// PrepareReader ...
func PrepareReader(r io.ReadSeeker, name string) (entry *Entry, err error) {

	entry, err = NewEntryReader(r, name)
	if err != nil {
		return
	}
	return
}

// PrepareFile ...
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

	return PrepareReader(f, name)
}

// ParseTags ...
func ParseTags(s string) (cdb.StringArray, error) {
	return strings.Split(strings.ToLower(s), ","), nil
}

// Delete ...
func Delete(roof, id string) error {
	if roof == "" {
		return ErrEmptyRoof
	}
	if id == "" {
		return ErrEmptyID
	}

	mw := NewMetaWrapper(roof)
	eid, err := imid.ParseID(id)
	if err != nil {
		return err
	}
	err = mw.Delete(eid.String())
	if err != nil {
		return err
	}
	return nil
}

// GetURI ...
func GetURI(suffix string) string {
	spath := path.Join("/", CatView, suffix)
	stageHost := config.Current.StageHost
	if stageHost == "" {
		return spath
	}
	return fmt.Sprintf("//%s%s", stageHost, spath)
}
