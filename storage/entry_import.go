package storage

import (
	"errors"
	"log"
	"time"

	"github.com/go-imsto/imagid"
	"github.com/go-imsto/imsto/image"
	"github.com/go-imsto/imsto/storage/types"
)

func NewEntryConvert(id, name, path string, size uint32, meta, sev types.Meta, hashes, ids []string,
	created time.Time) (entry *Entry, err error) {

	if id == "" {
		err = ErrEmptyID
		return
	}

	var eid imagid.IID
	eid, err = imagid.ParseID(id)
	if err != nil {
		log.Println(err)
		return
	}

	if path == "" {
		err = errors.New("'path' is empty")
		return
	}

	if size == 0 {
		err = errors.New("zero 'size'")
		return
	}

	var ia image.Attr
	ia.FromMap(meta)

	entry = &Entry{
		Id:      eid,
		Name:    name,
		Path:    path,
		Size:    size,
		sev:     sev,
		Created: created,
	}
	ia.Name = name

	entry.Meta = &ia
	entry.Hashes = types.StringArray(hashes)

	entry.IDs = types.StringArray(ids)
	return
}
