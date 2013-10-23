package imsto

import (
	"crypto/md5"
	"fmt"
	"imsto/base"
	"imsto/image"
	"io"
	"io/ioutil"
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

type EntryName string

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

func NewEntryByReader(r io.Reader) (entry *Entry, err error) {
	var (
		buf  []byte
		hash string
		id   *EntryId
	)

	buf, err = ioutil.ReadAll(r)

	if err != nil {
		return empty_item, err
	}
	hash = fmt.Sprintf("%x", md5.Sum(buf))
	id, err = NewEntryIdFromHash(hash)

	if err != nil {
		return empty_item, err
	}

	it := image.GuessType(&buf)
	ext := image.ExtByType(it)
	path := newPath(id, ext)

	entry = &Entry{Id: id, Size: uint32(len(buf)), Path: path, Hashes: []string{hash}, Ids: []EntryId{*id}}

	return
}

func newPath(ei *EntryId, ext string) string {
	r := []byte(ei.id)
	p := string(r[0:2]) + "/" + string(r[2:4]) + "/" + string(r[4:]) + ext

	return p
}
