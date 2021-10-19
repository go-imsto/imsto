// imsto core objects
package storage

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"path"
	"time"

	iimg "github.com/go-imsto/imagi"
	"github.com/go-imsto/imid"
	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/storage/backend"
	"github.com/go-imsto/imsto/storage/hash"
	cdb "github.com/go-imsto/imsto/storage/types"
	"github.com/go-imsto/imsto/utils"
)

type IID = imid.IID
type AppID uint16

type Author uint32

type StringArray = cdb.StringArray
type StringSlice = cdb.StringSlice

type mapItem struct {
	ID      IID         `json:"id"`
	Hash    string      `json:"hash"`
	Name    string      `json:"name"`
	Size    uint32      `json:"size"`
	Path    string      `json:"path"`
	Roofs   StringArray `json:"roofs,omitempty"`
	Status  uint8       `json:"-"`
	Created *time.Time  `json:"created,omitempty"`
	sev     cdb.Meta
}

func (e *mapItem) roof() string {
	if len(e.Roofs) > 0 {
		return e.Roofs[0]
	}
	return ""
}

type Entry struct {
	Id      imid.IID    `json:"id"`
	Name    string      `json:"name"`
	Size    uint32      `json:"size"`
	Path    string      `json:"path"`
	Status  uint8       `json:"-"`
	Hashes  cdb.Meta    `json:"hashes,omitempty"`
	IDs     StringArray `json:"-"`
	Roofs   StringArray `json:"roofs,omitempty"`
	Tags    StringArray `json:"tags,omitempty"`
	Meta    *iimg.Attr  `json:"meta,omitempty"`
	AppId   AppID       `json:"appid,omitempty"`
	Author  Author      `json:"author,omitempty"`
	Created time.Time   `json:"created,omitempty"`

	Key string `json:"key,omitempty"`   // for upload response
	Err string `json:"error,omitempty"` // for upload response

	exif cdb.Meta
	sev  cdb.Meta

	b  []byte
	h  string
	im *iimg.Image

	_treked bool
	ret     int // db saved result
}

const (
	minSize = 43
)

func (e *Entry) GetHash() string {
	return e.h
}

// NewEntryReader ...
func NewEntryReader(rs io.ReadSeeker, name string) (e *Entry, err error) {
	w := hash.New()
	io.Copy(w, rs)
	e = &Entry{
		Name:    name,
		Created: time.Now(),
	}
	rs.Seek(0, 0)
	e.im, err = iimg.Open(rs)
	if err != nil {
		logger().Infow("open image fail", "name", name, "len", w.Len())
		return
	}
	e.Size = w.Len()
	e.Meta = e.im.Attr

	e.h = w.String()
	e.Tags = StringArray{}
	return
}

// Trek 处理图片信息并填充
func (e *Entry) Trek(roof string) (err error) {
	if e._treked {
		return
	}
	e._treked = true

	var wopt *iimg.WriteOption
	wopt, err = filterImageAttr(roof, e.im.Attr)
	if err != nil {
		return
	}

	var buf bytes.Buffer
	err = e.im.SaveTo(&buf, *wopt)
	if err != nil {
		logger().Infow("im.SaveTo fail", "id", e.Id, "err", err)
		return
	}
	logger().Infow("im.SaveTo OK", "id", e.Id, "size", buf.Len(), "name", e.Name)

	e.b = buf.Bytes()

	size := len(e.b)
	if uint(size) > config.Current.MaxFileSize {
		err = fmt.Errorf("file: %s size %d is too big, max is %d", e.Name, size, config.Current.MaxFileSize)
		return
	}

	hashes := cdb.Meta{"hash": e.h, "size": e.Size}
	ids := cdb.StringArray{e.Id.String()}
	hash2 := hash.SumContent(e.b)
	if hash2 != e.h {
		logger().Infow("hashed", "hash1", e.h, "hash2", hash2)
		hashes["hash2"] = hash2
		hashes["size2"] = size

		e.h = hash2
		e.Size = uint32(size)
		e.Meta = e.im.Attr
		e.Path = e.Id.String() + e.im.Attr.Ext
	}
	e.Hashes = hashes
	e.IDs = ids

	return
}

// Store ...
func (e *Entry) Store(roof string) (ch chan error) {
	ch = make(chan error, 1)
	// TODO: refactory
	if len(roof) == 0 {
		ch <- ErrEmptyRoof
		return
	}
	if v := config.GetEngine(roof); len(v) == 0 {
		ch <- ErrInvalidRoof
		return
	}
	mw := NewMetaWrapper(roof)
	if eh, err := mw.GetHash(e.h); err == nil {
		logger().Infow("exist hash", "eh", eh)

		e.Id = eh.ID
		e.Path = eh.Path
		_ne, _err := mw.GetMapping(eh.ID.String())
		if _err != nil {
			logger().Warnw("exist mapping is invalid", "ne", _ne, "err", _err)
			ch <- _err
			return
		}

		e.Name = _ne.Name
		// e.Path = _ne.Path
		e.Size = _ne.Size
		e.Created = *_ne.Created
		e.Roofs = _ne.Roofs
		e.sev = _ne.sev
		e.reset()
		e._treked = true

		if err = mw.Save(e, true); err != nil {
			logger().Warnw("mw.Save fail", "entry", e, "err", err)
			ch <- err
			return
		}
		logger().Infow("exist entry", "id", e.Id, "path", e.Path)
		close(ch)
		return
	}

	if e.Id == 0 || e.Path == "" {
		id, err := mw.NextID() // generate new ID
		if err != nil {
			logger().Infow("gen ID fail", "name", e.Name, "len", e.Size)
			return
		}
		e.Id = imid.IID(id)
		e.Path = e.Id.String() + e.im.Attr.Ext
	}

	if err := e.Trek(roof); err != nil {
		ch <- err
		return
	}
	logger().Infow("trek ok", "entry", e)
	// log.Printf("new id: %v, size: %d, path: %v\n", e.Id, e.Size, e.Path)

	thumbRoot := path.Join(config.Current.CacheRoot, "thumb")
	filename := path.Join(thumbRoot, "orig", storedPath(e.Path))

	if err := utils.SaveFile(filename, e.b); err != nil {
		logger().Infow("entry save file fail", "filename", filename)
		ch <- err
		return
	}

	if err := mw.Ready(e); err != nil {
		ch <- err
		return
	}

	go func() {
		if err := e._save(roof); err != nil {
			log.Printf("_save error: %s", err)
			ch <- err
		} else {
			close(ch)
		}
	}()

	log.Printf("[%s] store ready ok %s", roof, e.Path)

	return
}

func (e *Entry) _save(roof string) (err error) {
	en := config.GetEngine(roof)
	log.Printf("start save %s to engine %s", e.Id, en)

	e.sev, err = e.PushTo(roof)
	if err != nil {
		log.Printf("engine push error: %s", err)
		return
	}
	log.Printf("engine push %s ok", e.Id)

	mw := NewMetaWrapper(roof)
	if err = mw.SetDone(e.Id.String(), e.sev); err != nil {
		logger().Infow("setDone fail", "entry", e)
		// if err = mw.Save(e); err != nil {
		// 	return
		// }
		return
	}

	log.Printf("%s set done ok", e.Id)
	return
}

func (e *Entry) reset() {
	e.b = []byte{}
}

func (e *Entry) origFullname() string {
	thumbRoot := path.Join(config.Current.CacheRoot, "thumb")
	return path.Join(thumbRoot, "orig", storedPath(e.Path))
}

func (e *Entry) roof() string {
	if len(e.Roofs) > 0 {
		return e.Roofs[0]
	}
	return ""
}

// StoredMeta ...
func (e *Entry) StoredMeta() cdb.Meta {
	return e.sev
}

// URI ..
func (e *Entry) URI(sizeOp string) string {
	return GetURI(sizeOp + "/" + e.Path)
}

func getItemCat(roof string) string {
	if cat := config.GetPrefix(roof); cat != "" {
		return cat
	}
	return CatStore
}

// pullWith pull blob from engine with key path
func (e *mapItem) pullWith(roof string) (data []byte, err error) {
	logger().Infow("pulling", "roof", roof, "path", e.Path)
	var em backend.Wagoner
	em, err = backend.FarmEngine(roof)
	if err != nil {
		logger().Warnw("farmEngine fail", "roof", roof, "err", err)
		return
	}
	cat := getItemCat(roof)
	if v, ok := e.sev.Get("cat"); ok {
		if s, ok2 := v.(string); ok2 {
			cat = s
		}
	}
	// var data []byte
	key := backend.Key{ID: e.Path, Cat: cat}
	data, err = em.Get(key)
	if err != nil {
		logger().Warnw("get fail", "roof", roof, "key", key, "err", err)
		return
	}
	logger().Infow("pulled", "roof", roof, "bytes", len(data))
	return
}

// PushTo ...
func (e *Entry) PushTo(roof string) (sev cdb.Meta, err error) {
	key := e.Path
	blob := e.b
	meta := e.Meta
	if e.im != nil {
		meta = e.im.Attr
	}
	var em backend.Wagoner
	em, err = backend.FarmEngine(roof)
	if err != nil {
		logger().Warnw("farmEngine fail", "roof", roof, "key", key, "err", err)
		return
	}

	sev, err = em.Put(backend.Key{ID: key, Cat: getItemCat(roof)}, blob, meta.ToMap())
	return
}

func filterImageAttr(roof string, ia *iimg.Attr) (wopt *iimg.WriteOption, err error) {

	maxQuality := iimg.Quality(config.Current.MaxQuality)
	if ia.Quality > 0 && maxQuality > 0 {
		if ia.Quality > maxQuality {
			log.Printf("jpeg quality %d is too high, set to %d", ia.Quality, maxQuality)
		} else {
			maxQuality = ia.Quality
			log.Printf("jpeg quality %d is too low", ia.Quality)
		}
	}

	maxWidth := iimg.Dimension(config.Current.MaxWidth)
	maxHeight := iimg.Dimension(config.Current.MaxHeight)
	if ia.Width > maxWidth || ia.Height > maxHeight {
		logger().Infow("dimension warning", "maxWidth", maxWidth, "maxHeight", maxHeight, "ia", ia)
		err = fmt.Errorf("dimension %dx%d of %s is too big", ia.Width, ia.Height, ia.Ext)
		return
	}

	minWidth := iimg.Dimension(config.Current.MinWidth)
	minHeight := iimg.Dimension(config.Current.MinHeight)
	if ia.Width < minWidth || ia.Height < minHeight {
		err = fmt.Errorf("dimension %dx%d of %s is too small", ia.Width, ia.Height, ia.Ext)
		return
	}

	wopt = &iimg.WriteOption{Quality: maxQuality}
	return
}
