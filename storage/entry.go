package storage

import (
	"bytes"
	"calf/base"
	"calf/image"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
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

func NewEntryIdById(id string) (*EntryId, error) {
	hash, err := base.BaseConvert(id, 36, 16)
	return &EntryId{id, hash}, err
}

func (ei *EntryId) String() string {
	return ei.id
}

type AppId uint16

type Author uint16

type Mime string

type ImageAttr struct {
	Width   uint32 // image width
	Height  uint32 // image height
	Quality uint8  // image compression quality
	Format  string // image format, like 'JPEG', 'PNG'
}

type Entry struct {
	Id        *EntryId
	Name      string
	Hashes    []string
	Ids       []EntryId
	Meta      *ImageAttr
	Size      uint32
	AppId     AppId
	Author    Author
	Path      string
	imageType int
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

	if f, ok := r.(*os.File); ok {
		f.Seek(0, 0)
	} else if rr, ok := r.(*bytes.Buffer); ok {
		rr.Reset()
	}

	im, err = image.Open(r)

	if err != nil {
		fmt.Println(err)
		return empty_item, err
	}

	defer im.Close()

	ia := im.GetAttr()

	var size uint
	data := im.Blob(&size)

	hash = fmt.Sprintf("%x", md5.Sum(buf))
	id2, err = NewEntryIdFromHash(hash)

	if err != nil {
		fmt.Println(err)
		return empty_item, err
	}

	ext := ia.Ext
	path := newPath(id, ext)

	entry = &Entry{Id: id, Size: uint32(size), Path: path, Hashes: []string{hash}, Ids: []EntryId{*id}}

	return
}

func newPath(ei *EntryId, ext string) string {
	r := []byte(ei.id)
	p := string(r[0:2]) + "/" + string(r[2:4]) + "/" + string(r[4:]) + ext

	return p
}
