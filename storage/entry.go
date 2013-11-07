package storage

import (
	"bytes"
	"calf/base"
	cdb "calf/db"
	"calf/image"
	"crypto/md5"
	// "errors"
	"errors"
	"fmt"
	// "io"
	// "io/ioutil"
	"log"
	"mime"
	// "os"
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

// func (ei *EntryId) MarshalJSON() ([]byte, error) {
// 	return []byte(ei.id), nil
// }

func (ei *EntryId) Hashed() string {
	return ei.hash
}

func (ei *EntryId) tip() string {
	return ei.id[:1]
}

type AppId uint16

type Author uint16

// type ImageAttr struct {
// 	Width   uint32 // image width
// 	Height  uint32 // image height
// 	Quality uint8  // image compression quality
// 	Format  string // image format, like 'JPEG', 'PNG'
// }

type Entry struct {
	Id        *EntryId //`json:"id,omitempty"`
	Name      string
	Hashes    cdb.Qarray
	Ids       cdb.Qarray
	Meta      *image.ImageAttr
	Size      uint32
	AppId     AppId
	Author    Author
	Path      string
	Mime      string
	imageType int
	sev       cdb.Hstore
}

var empty_item = &Entry{}

const (
	min_size = 43
)

func NewEntry(data []byte) (entry *Entry, err error) {
	if len(data) < min_size {
		err = errors.New("data is too small, maybe not a valid image")
		return
	}
	var (
		hash string
		id   *EntryId
		im   image.Image
	)

	rd := bytes.NewReader(data)
	im, err = image.Open(rd)

	if err != nil {
		log.Println(err)
		return empty_item, err
	}

	defer im.Close()

	hash = HashContent(data)
	id, err = NewEntryIdFromHash(hash)

	hashes := cdb.Qarray{hash}
	ids := cdb.Qarray{id.String()}

	ia := im.GetAttr()
	// log.Println(ia)
	var size uint
	data = im.Blob(&size) // tack new data

	// TODO: 添加最小优化比率判断，如果过小，就忽略

	var hash2 string
	hash2 = HashContent(data)
	if hash2 != hash {
		hashes = append(hashes, hash2)
		var id2 *EntryId
		id2, err = NewEntryIdFromHash(hash2)
		ids = append(ids, id2.String())
		id = id2 // 使用新的 Id 作为主键
	}

	if err != nil {
		log.Println(err)
		return empty_item, err
	}

	ia.Size = uint32(size) // 更新后的大小

	ext := ia.Ext
	path := newPath(id, ext)
	mimetype := mime.TypeByExtension(ext)
	ia.Mime = mimetype

	log.Printf("ext: %s, mime: %s\n", ext, mimetype)

	entry = &Entry{Id: id, Name: "", Size: ia.Size, Meta: ia, Path: path, Mime: mimetype, Hashes: hashes, Ids: ids}

	return
}

func newPath(ei *EntryId, ext string) string {
	r := []byte(ei.id)
	p := string(r[0:2]) + "/" + string(r[2:4]) + "/" + string(r[4:]) + ext

	return p
}

func HashContent(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}
