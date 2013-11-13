package storage

import (
	"bytes"
	"calf/base"
	"calf/config"
	cdb "calf/db"
	iimg "calf/image"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
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

type AppId uint16

type Author uint16

type Entry struct {
	Id        *EntryId        `json:"id,omitempty"`
	Name      string          `json:"name"`
	Hashes    cdb.Qarray      `json:"-"`
	Ids       cdb.Qarray      `json:"-"`
	Meta      *iimg.Attr `json:"meta,omitempty"`
	Size      uint32          `json:"size"`
	AppId     AppId           `json:"-"`
	Author    Author          `json:"-"`
	Path      string          `json:"path"`
	Mime      string          `json:"mime"`
	Modified  uint64          `json:"modified"`
	Created   time.Time       `json:"created"`
	imageType int
	sev       cdb.Hstore
	exif      cdb.Hstore
	b         []byte
	h         string
	_treked   bool
}

const (
	min_size = 43
)

func newEntry(data []byte, name string) (e *Entry, err error) {
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
		Id:   id,
		Name: name,
		Size: uint32(len(data)),
		b:    data,
		h:    hash,
	}

	// entry = &Entry{Id: id, Name: name, Size: ia.Size, Meta: ia, Path: path, Mime: mimetype, Hashes: hashes, Ids: ids}

	return
}

// 处理图片信息并填充
func (e *Entry) Trek(section string) (err error) {
	if e._treked {
		return
	}
	e._treked = true
	var im iimg.Image
	rd := bytes.NewReader(e.b)
	im, err = iimg.Open(rd)

	if err != nil {
		log.Println(err)
		return
	}

	defer im.Close()

	hashes := cdb.Qarray{e.h}
	ids := cdb.Qarray{e.Id.String()}

	ia := im.GetAttr()
	// log.Println(ia)

	max_quality := iimg.Quality(config.GetInt(section, "max_quality"))
	if ia.Quality > max_quality {
		im.SetOption(iimg.WriteOption{Quality: max_quality, StripAll: true})
		log.Printf("set quality to max_quality %d", max_quality)
	}

	var data []byte
	data, err = im.GetBlob() // tack new data

	if err != nil {
		log.Println(err)
		return
	}

	// TODO: 添加最小优化比率判断，如果过小，就忽略

	var hash2 string
	size := len(data)
	if max_file_size := config.GetInt(section, "max_file_size"); size > max_file_size {
		err = errors.New(fmt.Sprintf("file: %s size %d is too big, max is %d", e.Name, size, max_file_size))
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

	path := newPath(e.Id, ia.Ext)

	log.Printf("ext: %s, mime: %s\n", ia.Ext, ia.Mime)

	e.Meta = ia
	e.Path = path
	e.Mime = ia.Mime
	e.Hashes = hashes
	e.Ids = ids
	return
}

func (e *Entry) Hashed() string {
	return e.h
}

func (e *Entry) Blob() []byte {
	return e.b
}

func newPath(ei *EntryId, ext string) string {
	r := []byte(ei.id)
	p := string(r[0:2]) + "/" + string(r[2:4]) + "/" + string(r[4:]) + ext

	return p
}

func HashContent(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}
