// imsto core objects
package storage

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"path"
	"time"
	"wpst.me/calf/config"
	cdb "wpst.me/calf/db"
	iimg "wpst.me/calf/image"
	"wpst.me/calf/storage/backend"
)

type AppId uint16

type Author uint32

type Entry struct {
	Id       *EntryId   `json:"id"`
	Name     string     `json:"name"`
	Size     uint32     `json:"size"`
	Path     string     `json:"path"`
	Mime     string     `json:"-"`
	Status   uint8      `json:"-"`
	Hashes   cdb.Qarray `json:"-"`
	Ids      cdb.Qarray `json:"-"`
	Roofs    cdb.Qarray `json:"roofs,omitempty"`
	Tags     cdb.Qarray `json:"tags,omitempty"`
	Meta     *iimg.Attr `json:"meta,omitempty"`
	AppId    AppId      `json:"appid,omitempty"`
	Author   Author     `json:"author,omitempty"`
	Modified uint64     `json:"modified,omitempty"`
	Created  time.Time  `json:"created,omitempty"`
	exif     cdb.Hstore
	sev      cdb.Hstore
	b        []byte
	h        string
	_treked  bool
	ret      int       // db saved result
	Done     chan bool `json:"-"`
	ready    int
}

const (
	min_size = 43
)

func NewEntry(data []byte, name string) (e *Entry, err error) {
	if len(data) < min_size {
		err = errors.New("data is too small, maybe not a valid image")
		return
	}

	var id *EntryId
	id, err = NewEntryIdFromData(data)

	if err != nil {
		log.Println(err)
		return
	}

	e = &Entry{
		Id:      id,
		Name:    name,
		Size:    uint32(len(data)),
		Created: time.Now(),
		b:       data,
		h:       id.hash,
	}

	// entry = &Entry{Id: id, Name: name, Size: ia.Size, Meta: ia, Path: path, Mime: mimetype, Hashes: hashes, Ids: ids}

	return
}

// 处理图片信息并填充
func (e *Entry) Trek(roof string) (err error) {
	if e._treked {
		return
	}
	e._treked = true
	var im iimg.Image
	rd := bytes.NewReader(e.b)
	im, err = iimg.Open(rd)

	if err != nil {
		log.Printf("image open error: %s", err)
		return
	}

	defer im.Close()

	ia := im.GetAttr()
	// log.Println(ia)

	max_quality := iimg.Quality(config.GetInt(roof, "max_quality"))
	if ia.Quality > max_quality {
		log.Printf("jpeg quality %d is too high, set to %d", ia.Quality, max_quality)
	} else {
		max_quality = ia.Quality
		log.Printf("jpeg quality %d is too low", ia.Quality)
	}
	im.SetOption(iimg.WriteOption{Quality: max_quality, StripAll: true})

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

	var data []byte
	data, err = im.GetBlob() // tack new data

	if err != nil {
		log.Printf("GetBlob error: %s", err)
		return
	}

	hashes := cdb.Qarray{e.h}
	ids := cdb.Qarray{e.Id.String()}

	size := len(data)
	if max_file_size := config.GetInt(roof, "max_file_size"); size > max_file_size {
		err = fmt.Errorf("file: %s size %d is too big, max is %d", e.Name, size, max_file_size)
		return
	}

	var id2 *EntryId
	id2, err = NewEntryIdFromData(data)
	if id2.hash != e.h {
		hashes = append(hashes, id2.hash)
		if err != nil {
			// log.Println(err)
			return
		}
		ids = append(ids, id2.id)
		e.Id = id2 // 使用新的 Id 作为主键
		e.h = id2.hash
		e.b = data
		e.Size = uint32(size)
	}

	ia.Size = iimg.Size(size) // 更新后的大小
	ia.Name = e.Name

	path := newUrlPath(e.Id, ia.Ext)

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
		log.Printf("check hash error: %s", _err)
	} else if eh != nil && eh.id != "" {
		if _id, _err := NewEntryId(eh.id); _err == nil {
			e.Id = _id
			_ne, _err := mw.GetEntry(*_id)
			if _err == nil { // path, mime, size, sev, status, created
				if _ne.Roofs.Contains(roof) {
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

					log.Printf("exist: %s, %s", e.Id, e.Path)
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
	log.Printf("start save to engine %s", en)

	e.sev, err = PushBlob(e, roof)
	if err != nil {
		log.Printf("engine push error: %s", err)
		return
	}
	log.Print("engine push ok")

	mw := NewMetaWrapper(roof)
	if err = mw.SetDone(*e.Id, e.sev); err != nil {
		log.Println(err)
		// if err = mw.Save(e); err != nil {
		// 	return
		// }
		return
	}
	e.ready = -1
	log.Print("meta set done ok")
	return
}

func (e *Entry) fill(data []byte) error {
	if size := len(data); size != int(e.Size) {
		return fmt.Errorf("invliad size: %d (%d)", size, e.Size)
	}

	_, m := HashContent(data)

	if !e.Hashes.Contains(m) {
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
		return e.Roofs[0].(string)
	}
	return ""
}

func newUrlPath(ei *EntryId, ext string) string {
	return ei.id + ext
}

func newStorePath(ei *EntryId, ext string) string {
	r := ei.id
	p := r[0:2] + "/" + r[2:4] + "/" + r[4:] + ext

	return p
}

func PullBlob(e *Entry, roof string) (data []byte, err error) {
	var em backend.Wagoner
	em, err = backend.FarmEngine(roof)
	if err != nil {
		log.Printf("FarmEngine(%s) error: %s", roof, err)
		return
	}
	// var data []byte
	data, err = em.Get(e.Path)
	if err != nil {
		log.Printf("[%s] engine Get(%s) error: %s", roof, e.Path, err)
	}
	return
}

func PushBlob(e *Entry, roof string) (sev cdb.Hstore, err error) {
	var em backend.Wagoner
	em, err = backend.FarmEngine(roof)
	if err != nil {
		log.Printf("farm engine error: %s", err)
		return
	}
	sev, err = em.Put(e.Path, e.Blob(), e.Meta.Hstore())
	return
}
