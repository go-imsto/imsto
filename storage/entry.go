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
	"strconv"
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
	Meta      *iimg.ImageAttr `json:"meta,omitempty"`
	Size      uint32          `json:"size"`
	AppId     AppId           `json:"-"`
	Author    Author          `json:"-"`
	Path      string          `json:"path"`
	Mime      string          `json:"mime"`
	Modified  uint64          `json:"modified"`
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

	mq, _ := strconv.ParseUint(config.GetValue(section, "max_quality"), 10, 8)

	max_quality := iimg.Quality(mq)
	// log.Printf("max_quality: %d\b", max_quality)
	if ia.Quality > max_quality {
		im.SetOption(iimg.WriteOption{Quality: max_quality, StripAll: true})
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
	hash2 = HashContent(data)
	if hash2 != e.h {
		hashes = append(hashes, hash2)
		var id2 *EntryId
		id2, err = NewEntryIdFromHash(hash2)
		ids = append(ids, id2.String())
		e.Id = id2 // 使用新的 Id 作为主键
		e.h = hash2
		e.b = data
		e.Size = uint32(size)
	}

	if err != nil {
		log.Println(err)
		return
	}

	ia.Size = iimg.Size(size) // 更新后的大小

	// ext := ia.Ext
	path := newPath(e.Id, ia.Ext)
	// mimetype := mime.TypeByExtension(ext)
	// ia.Mime = mimetype

	log.Printf("ext: %s, mime: %s\n", ia.Ext, ia.Mime)

	// entry = &Entry{Id: id, Name: name, Size: ia.Size, Meta: ia, Path: path, Mime: mimetype, Hashes: hashes, Ids: ids}
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
