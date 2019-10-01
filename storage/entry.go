// imsto core objects
package storage

import (
	"bytes"
	"errors"
	"fmt"
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

type Entry struct {
	Id       base.PinID  `json:"id"`
	Name     string      `json:"name"`
	Size     uint32      `json:"size"`
	Path     string      `json:"path"`
	Mime     string      `json:"-"`
	Status   uint8       `json:"-"`
	Hashes   StringArray `json:"-"`
	Ids      StringArray `json:"-"`
	Roofs    StringArray `json:"roofs,omitempty"`
	Tags     StringArray `json:"tags,omitempty"`
	Meta     *iimg.Attr  `json:"meta,omitempty"`
	AppId    AppID       `json:"appid,omitempty"`
	Author   Author      `json:"author,omitempty"`
	Modified uint64      `json:"modified,omitempty"`
	Created  time.Time   `json:"created,omitempty"`
	exif     cdb.JsonKV
	sev      cdb.JsonKV
	b        []byte
	h        string
	_treked  bool
	ret      int       // db saved result
	Done     chan bool `json:"-"`
	ready    int
}

const (
	minSize = 43
)

// NewEntry
func NewEntry(data []byte, name string) (e *Entry, err error) {
	if len(data) < minSize {
		err = errors.New("data is too small, maybe not a valid image")
		return
	}

	id, hash := HashContent(data)
	pinID := base.PinID(id)

	e = &Entry{
		Id:      pinID,
		Name:    name,
		Size:    uint32(len(data)),
		Created: time.Now(),
		b:       data,
		h:       hash,
	}

	// entry = &Entry{Id: id, Name: name, Size: ia.Size, Meta: ia, Path: path, Mime: mimetype, Hashes: hashes, Ids: ids}

	return
}

// Trek 处理图片信息并填充
func (e *Entry) Trek(roof string) (err error) {
	if e._treked {
		return
	}
	e._treked = true
	var im *iimg.Image
	rd := bytes.NewReader(e.b)
	im, err = iimg.Open(rd, e.Name)

	if err != nil {
		log.Printf("image open error: %s", err)
		return
	}

	ia := im.Attr

	max_quality := iimg.Quality(config.GetInt(roof, "max_quality"))
	if ia.Quality > max_quality {
		log.Printf("jpeg quality %d is too high, set to %d", ia.Quality, max_quality)
	} else {
		max_quality = ia.Quality
		log.Printf("jpeg quality %d is too low", ia.Quality)
	}
	// im.SetOption(iimg.WriteOption{Quality: max_quality, StripAll: true})

	max_width := iimg.Dimension(config.GetInt(roof, "max_width"))
	max_height := iimg.Dimension(config.GetInt(roof, "max_height"))
	if ia.Width > max_width || ia.Height > max_height {
		err = fmt.Errorf("dimension %dx%d of %s is too big", ia.Width, ia.Height, e.Name)
		return
	}

	min_width := iimg.Dimension(config.GetInt(roof, "min_width"))
	min_height := iimg.Dimension(config.GetInt(roof, "min_height"))
	if ia.Width < min_width || ia.Height < min_height {
		err = fmt.Errorf("dimension %dx%d of %s is too small", ia.Width, ia.Height, e.Name)
		return
	}

	var buf bytes.Buffer
	err = im.WriteTo(&buf, &iimg.WriteOption{Quality: max_quality, StripAll: true})
	if err != nil {
		return
	}
	data := buf.Bytes()

	if err != nil {
		log.Printf("GetBlob error: %s", err)
		return
	}

	hashes := cdb.StringArray{e.h}
	ids := cdb.StringArray{e.Id.String()}

	size := len(data)
	if maxFileSize := config.GetInt(roof, "max_file_size"); size > maxFileSize {
		err = fmt.Errorf("file: %s size %d is too big, max is %d", e.Name, size, maxFileSize)
		return
	}

	id2, hash2 := HashContent(data)
	if hash2 != e.h {
		hashes = append(hashes, hash2)
		if err != nil {
			// log.Println(err)
			return
		}

		ids = append(ids, PinID(id2).String())
		e.Id = PinID(id2) // 使用新的 Id 作为主键
		e.h = hash2
		e.b = data
		e.Size = uint32(size)
	}

	ia.Size = iimg.Size(size) // 更新后的大小
	ia.Name = e.Name

	path := e.Id.String() + ia.Ext

	log.Printf("ext: %s, mime: %s\n", ia.Ext, ia.Mime)

	e.Meta = ia
	e.Path = path
	e.Mime = ia.Mime
	e.Hashes = hashes
	e.Ids = ids
	return
}

// return hash value string
// func (e *Entry) Hashed() string {
// 	return e.h
// }
func (e *Entry) storedPath() string {
	r := e.Id.String()
	ext := path.Ext(e.Path)
	p := r[0:2] + "/" + r[2:4] + "/" + r[4:] + ext

	return p
}

// return binary bytes
func (e *Entry) Blob() []byte {
	return e.b
}

func (e *Entry) IsDone() bool {
	return e.ready != 1
}

func (e *Entry) Store(roof string) (err error) {

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
					e.Mime = _ne.Mime
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

				// for _, _roof := range _ne.Roofs {
				// 	_mw := NewMetaWrapper(fmt.Sprint(_roof))
				// 	t, te := _mw.GetMeta(*_ne.Id)
				// 	if te == nil {
				// 		e = t
				// 		err = mw.Save(t)
				// 		return
				// 	}
				// }

				// e.Done <- true
			} else {
				log.Printf("get entry error: %s", _err)
			}
		}
	}

	if err = e.Trek(roof); err != nil {
		return
	}
	log.Printf("new id: %v, size: %d, path: %v\n", e.Id, e.Size, e.Path)

	data := e.Blob()
	// size := len(data)
	// log.Printf("blob length: %d", size)
	thumb_root := config.GetValue(roof, "thumb_root")
	filename := path.Join(thumb_root, "orig", e.storedPath())
	err = SaveFile(filename, data)
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

	e.sev, err = PushBlob(roof, e.Path, e.Blob(), e.Meta)
	if err != nil {
		log.Printf("engine push error: %s", err)
		return
	}
	log.Printf("engine push %s ok", e.Id)

	mw := NewMetaWrapper(roof)
	if err = mw.SetDone(e.Id.String(), e.sev); err != nil {
		log.Println(err)
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
	return path.Join(thumb_root, "orig", e.storedPath())
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

func PushBlob(roof, key string, blob []byte, meta *iimg.Attr) (sev cdb.JsonKV, err error) {
	var em backend.Wagoner
	em, err = backend.FarmEngine(roof)
	if err != nil {
		logger().Warnw("farmEngine fail", "roof", roof, "key", key, "err", err)
		return
	}
	sev, err = em.Put(key, blob, meta.ToMap())
	return
}
