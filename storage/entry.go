package storage

import (
	"bytes"
	"calf/base"
	"calf/image"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"os"
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
	Id        *EntryId
	Name      string
	Hashes    []string
	Ids       []string
	Meta      *image.ImageAttr
	Size      uint32
	AppId     AppId
	Author    Author
	Path      string
	Mime      string
	imageType int
	sev       hstore
}

var empty_item = &Entry{}

func NewEntry(r io.Reader) (entry *Entry, err error) {
	var (
		buf  []byte
		hash string
		id   *EntryId
		im   image.Image
	)

	buf, err = ioutil.ReadAll(r)

	if err != nil {
		return empty_item, err
	}
	hash = fmt.Sprintf("%x", md5.Sum(buf))
	id, err = NewEntryIdFromHash(hash)

	hashes := []string{hash}
	ids := []string{id.String()}

	if f, ok := r.(*os.File); ok {
		f.Seek(0, 0)
	} else if rr, ok := r.(*bytes.Buffer); ok {
		rr.Reset()
	}

	im, err = image.Open(r)

	if err != nil {
		log.Println(err)
		return empty_item, err
	}

	defer im.Close()

	ia := im.GetAttr()
	// log.Println(ia)
	var size uint
	data := im.Blob(&size)

	var hash2 string
	hash2 = fmt.Sprintf("%x", md5.Sum(data))
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

	entry = &Entry{Id: id, Name: "", Size: ia.Size, Meta: ia, Path: path, Mime: mimetype, Hashes: hashes, Ids: ids}

	// log.Println(ia2hstore(entry.Meta))
	return
}

func newPath(ei *EntryId, ext string) string {
	r := []byte(ei.id)
	p := string(r[0:2]) + "/" + string(r[2:4]) + "/" + string(r[4:]) + ext

	return p
}

func ia2hstore(ia image.KVMapper) (m hstore) {
	m = hstore(ia.Maps())
	return
}
