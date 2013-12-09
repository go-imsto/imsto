// imsto core objects
package storage

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
	"wpst.me/calf/base"
	"wpst.me/calf/config"
	cdb "wpst.me/calf/db"
	iimg "wpst.me/calf/image"
)

type EntryId struct {
	id   string
	hash string
}

func NewEntryIdFromHash(hash string) (*EntryId, error) {
	id, err := base.BaseConvert(hash, 16, 36)

	return &EntryId{id, hash}, err
}

func NewEntryId(id string) (*EntryId, error) {
	hash, err := base.BaseConvert(id, 36, 16)
	return &EntryId{id, hash}, err
}

func (ei *EntryId) String() string {
	return ei.id
}

func (ei *EntryId) MarshalJSON() ([]byte, error) {
	return json.Marshal(ei.id)
}

func (ei *EntryId) Hashed() string {
	return ei.hash
}

func (ei *EntryId) tip() string {
	return ei.id[:1]
}

type AppId uint8

type Author uint32

type Entry struct {
	Id        *EntryId   `json:"id,omitempty"`
	Name      string     `json:"name,omitempty"`
	Size      uint32     `json:"size"`
	Path      string     `json:"path"`
	Mime      string     `json:"mime,omitempty"`
	Status    uint8      `json:"-"`
	Hashes    cdb.Qarray `json:"-"`
	Ids       cdb.Qarray `json:"-"`
	Meta      *iimg.Attr `json:"meta,omitempty"`
	AppId     AppId      `json:"appid,omitempty"`
	Author    Author     `json:"author,omitempty"`
	Modified  uint64     `json:"modified,omitempty"`
	Created   time.Time  `json:"created,omitempty"`
	imageType int
	exif      cdb.Hstore
	sev       cdb.Hstore
	b         []byte
	h         string
	_treked   bool
	ret       int // db saved result
}

const (
	min_size = 43
)

func NewEntry(data []byte, name string) (e *Entry, err error) {
	if len(data) < min_size {
		err = errors.New("data is too small, maybe not a valid image")
		return
	}

	hash := HashContent(data)
	var id *EntryId
	id, err = NewEntryIdFromHash(hash)

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
		h:       hash,
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
		im.SetOption(iimg.WriteOption{Quality: max_quality, StripAll: true})
		log.Printf("jpeg quality %d is too high, set to %d", ia.Quality, max_quality)
	}

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

	var hash2 string
	size := len(data)
	if max_file_size := config.GetInt(roof, "max_file_size"); size > max_file_size {
		err = fmt.Errorf("file: %s size %d is too big, max is %d", e.Name, size, max_file_size)
		return
	}

	hash2 = HashContent(data)
	if hash2 != e.h {
		hashes = append(hashes, hash2)
		var id2 *EntryId
		id2, err = NewEntryIdFromHash(hash2)
		if err != nil {
			// log.Println(err)
			return
		}
		ids = append(ids, id2.String())
		e.Id = id2 // 使用新的 Id 作为主键
		e.h = hash2
		e.b = data
		e.Size = uint32(size)
	}

	ia.Size = iimg.Size(size) // 更新后的大小
	ia.Name = e.Name

	path := newPath(e.Id, ia.Ext)

	log.Printf("ext: %s, mime: %s\n", ia.Ext, ia.Mime)

	e.Meta = ia
	e.Path = path
	e.Mime = ia.Mime
	e.Hashes = hashes
	e.Ids = ids
	return
}

// return hash value string
func (e *Entry) Hashed() string {
	return e.h
}

// return binary bytes
func (e *Entry) Blob() []byte {
	return e.b
}

func (e *Entry) store(roof string) (err error) {

	mw := NewMetaWrapper(roof)
	eh, _err := mw.GetHash(e.h)
	if _err != nil { // ok, not exsits
		log.Printf("check hash error: %s", _err)
	} else if eh != nil && eh.id != "" {
		if _id, _err := NewEntryId(eh.id); _err == nil {
			e.Id = _id
			_ne, _err := mw.GetEntry(*_id)
			if _err == nil { // path, mime, size, sev, status, created
				e.Name = _ne.Name
				e.Path = _ne.Path
				e.Size = _ne.Size
				e.Mime = _ne.Mime
				e.sev = _ne.sev
				e.Created = _ne.Created
				e.reset()
				e._treked = true
				log.Printf("exist: %s, %s", e.Id, e.Path)
				return
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

	var em Wagoner
	if em, err = FarmEngine(roof); err != nil {
		// log.Println(err)
		return
	}

	sev, dbe := em.Put(e.Path, data, e.Meta.Hstore())
	if dbe != nil {
		log.Printf("engine put error: %s", dbe)
		err = dbe
		return
	}

	e.sev = sev

	if err = mw.Save(e); err != nil {
		// log.Println(err)
		return
	}
	return
}

func (e *Entry) reset() {
	e.b = []byte{}
}

func newPath(ei *EntryId, ext string) string {
	r := ei.id
	p := r[0:2] + "/" + r[2:4] + "/" + r[4:] + ext

	return p
}

func HashContent(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}
