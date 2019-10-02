// imsto core objects
package storage

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"path"
	"time"

	"github.com/go-imsto/imsto/base"
	"github.com/go-imsto/imsto/config"
	iimg "github.com/go-imsto/imsto/image"
	"github.com/go-imsto/imsto/storage/backend"
	cdb "github.com/go-imsto/imsto/storage/types"
)

type PinID = base.PinID
type AppID uint16

type Author uint32

type StringArray = cdb.StringArray
type StringSlice = cdb.StringSlice

type mapItem struct {
	ID      PinID       `json:"id"`
	Hash    string      `json:"hash"`
	Name    string      `json:"name"`
	Size    uint32      `json:"size"`
	Path    string      `json:"path"`
	Roofs   StringArray `json:"roofs,omitempty"`
	Status  uint8       `json:"-"`
	Created *time.Time  `json:"created,omitempty"`
	sev     cdb.JsonKV
}

func (e *mapItem) roof() string {
	if len(e.Roofs) > 0 {
		return e.Roofs[0]
	}
	return ""
}

type Entry struct {
	Id       base.PinID  `json:"id"`
	Name     string      `json:"name"`
	Size     uint32      `json:"size"`
	Path     string      `json:"path"`
	Status   uint8       `json:"-"`
	Hashes   StringArray `json:"hashes,omitempty"`
	IDs      StringArray `json:"-"`
	Roofs    StringArray `json:"roofs,omitempty"`
	Tags     StringArray `json:"tags,omitempty"`
	Meta     *iimg.Attr  `json:"meta,omitempty"`
	AppId    AppID       `json:"appid,omitempty"`
	Author   Author      `json:"author,omitempty"`
	Modified uint64      `json:"modified,omitempty"`
	Created  time.Time   `json:"created,omitempty"`

	exif cdb.JsonKV
	sev  cdb.JsonKV

	b  []byte
	h  string
	im *iimg.Image

	_treked bool
	ret     int       // db saved result
	Done    chan bool `json:"-"`
	ready   int
}

const (
	minSize = 43
)

// NewEntryReader ...
func NewEntryReader(rs io.ReadSeeker, name string) (e *Entry, err error) {
	w := newHasher()
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
	e.im.Name = name
	e.Meta = e.im.Attr
	id, hash := w.Hash()
	e.Id = base.PinID(id)
	e.h = hash
	e.Path = e.Id.String() + e.im.Attr.Ext
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
	err = e.im.WriteTo(&buf, wopt)
	if err != nil {
		logger().Infow("im.WriteTo fail", "id", e.Id, "err", err)
		return
	}
	e.b = buf.Bytes()

	size := len(e.b)
	if maxFileSize := config.GetInt(roof, "max_file_size"); size > maxFileSize {
		err = fmt.Errorf("file: %s size %d is too big, max is %d", e.Name, size, maxFileSize)
		return
	}

	hashes := cdb.StringArray{e.h}
	ids := cdb.StringArray{e.Id.String()}
	id2, hash2 := HashContent(e.b)
	if hash2 != e.h {
		logger().Infow("hashed", "hash1", e.h, "hash2", hash2)
		hashes = append(hashes, hash2)

		ids = append(ids, PinID(id2).String())
		e.Id = PinID(id2) // 使用新的 Id 作为主键
		e.h = hash2
		e.Size = uint32(size)
		e.Meta = e.im.Attr
		e.Path = e.Id.String() + e.im.Attr.Ext
	}
	e.Hashes = hashes
	e.IDs = ids

	return
}

// return hash value string
// func (e *Entry) Hashed() string {
// 	return e.h
// }
// storedPath ...
func storedPath(r string) string {
	if len(r) < 5 {
		return r
	}
	p := r[0:2] + "/" + r[2:4] + "/" + r[4:]

	return p
}

func (e *Entry) IsDone() bool {
	return e.ready != 1
}

// Store ...
func (e *Entry) Store(roof string) (err error) {

	// TODO:
	mw := NewMetaWrapper(roof)
	eh, _err := mw.GetHash(e.h)
	if _err != nil { // ok, not exsits
		logger().Infow("check hash fail", "h", e.h, "err", _err)
	} else if eh != nil && eh.id != "" {
		if _id, _err := base.ParseID(eh.id); _err == nil {
			e.Id = _id
			_ne, _err := mw.GetMeta(_id.String())
			if _err == nil { // path, mime, size, sev, status, created
				if StringSlice(_ne.Roofs).Contains(roof) {
					e.Name = _ne.Name
					e.Path = _ne.Path
					e.Size = _ne.Size
					e.Meta = _ne.Meta
					// e.sev = _ne.sev
					e.Created = _ne.Created
					e.Roofs = _ne.Roofs
					e.reset()
					e._treked = true
					mw.Save(e, true)

					logger().Infow("exist entry", "id", e.Id, "path", e.Path)
					return
				}

				log.Printf("[%s]%s not in %s, so resubmit it", roof, e.Id, _ne.Roofs)

			} else {
				log.Printf("get entry error: %s", _err)
			}
		}
	}

	if err = e.Trek(roof); err != nil {
		return
	}
	logger().Infow("trek ok", "entry", e)
	// log.Printf("new id: %v, size: %d, path: %v\n", e.Id, e.Size, e.Path)

	thumbRoot := config.GetValue(roof, "thumb_root")
	filename := path.Join(thumbRoot, "orig", storedPath(e.Path))
	err = SaveFile(filename, e.b)
	if err != nil {
		return
	}

	err = mw.Ready(e)
	if err != nil {
		return
	}
	e.ready = 1

	e.Done = make(chan bool, 1)
	go func() {
		err = e._save(roof)
		if err != nil {
			log.Printf("_save error: %s", err)
		}
		e.Done <- true
	}()

	log.Printf("[%s] store ready ok %s", roof, e.Path)

	return
}

func (e *Entry) _save(roof string) (err error) {
	en := config.GetValue(roof, "engine")
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
	e.ready = -1
	log.Printf("%s set done ok", e.Id)
	return
}

func (e *Entry) fill(data []byte) error {
	if size := len(data); size != int(e.Size) {
		return fmt.Errorf("invliad size: %d (%d)", size, e.Size)
	}

	_, m := HashContent(data)

	if !StringSlice(e.Hashes).Contains(m) {
		return fmt.Errorf("invalid hash: %s (%s)", m, e.Hashes)
	}

	e.b = data
	return nil
}

func (e *Entry) reset() {
	e.b = []byte{}
}

func (e *Entry) origFullname() string {
	thumb_root := config.GetValue(e.roof(), "thumb_root")
	return path.Join(thumb_root, "orig", storedPath(e.Path))
}

func (e *Entry) roof() string {
	if len(e.Roofs) > 0 {
		return e.Roofs[0]
	}
	return ""
}

func newStorePath(id base.PinID, ext string) string {
	r := id.String()
	p := r[0:2] + "/" + r[2:4] + "/" + r[4:] + ext

	return p
}

// PullBlob pull blob from engine with key path
func PullBlob(key string, roof string) (data []byte, err error) {
	var em backend.Wagoner
	em, err = backend.FarmEngine(roof)
	if err != nil {
		logger().Warnw("farmEngine fail", "roof", roof, "err", err)
		return
	}
	// var data []byte
	data, err = em.Get(key)
	if err != nil {
		logger().Warnw("get fail", "roof", roof, "key", key, "err", err)
	}
	return
}

// PushTo ...
func (e *Entry) PushTo(roof string) (sev cdb.JsonKV, err error) {
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
	sev, err = em.Put(key, blob, meta.ToMap())
	return
}

func filterImageAttr(roof string, ia *iimg.Attr) (wopt *iimg.WriteOption, err error) {

	maxQuality := iimg.Quality(config.GetInt(roof, "max_quality"))
	if ia.Quality > 0 {
		if ia.Quality > maxQuality {
			log.Printf("jpeg quality %d is too high, set to %d", ia.Quality, maxQuality)
		} else {
			maxQuality = ia.Quality
			log.Printf("jpeg quality %d is too low", ia.Quality)
		}
	}

	maxWidth := iimg.Dimension(config.GetInt(roof, "max_width"))
	maxHeight := iimg.Dimension(config.GetInt(roof, "max_height"))
	if ia.Width > maxWidth || ia.Height > maxHeight {
		err = fmt.Errorf("dimension %dx%d of %s is too big", ia.Width, ia.Height, ia.Name)
		return
	}

	minWidth := iimg.Dimension(config.GetInt(roof, "min_width"))
	minHeight := iimg.Dimension(config.GetInt(roof, "min_height"))
	if ia.Width < minWidth || ia.Height < minHeight {
		err = fmt.Errorf("dimension %dx%d of %s is too small", ia.Width, ia.Height, ia.Name)
		return
	}

	wopt = &iimg.WriteOption{Quality: maxQuality, StripAll: true}
	return
}
