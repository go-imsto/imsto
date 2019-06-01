package storage

import (
	"errors"
	"github.com/go-imsto/imsto/image"
	"github.com/go-imsto/imsto/storage/types"
	"log"
	"time"
)

func NewEntryConvert(id, name, path, mime string, size uint32, meta, sev types.JsonKV, hashes, ids []string,
	created time.Time) (entry *Entry, err error) {

	if id == "" {
		err = errors.New("'id' is empty")
		return
	}

	var eid *EntryId
	eid, err = NewEntryId(id)

	if err != nil {
		log.Println(err)
		return
	}

	if path == "" {
		err = errors.New("'path' is empty")
		return
	}

	if mime == "" {
		err = errors.New("'mime' is empty")
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
		Mime:    mime,
		sev:     sev,
		Created: created,
	}
	ia.Size = image.Size(size)
	ia.Name = name
	if ia.Mime == "" && mime != "" {
		ia.Mime = mime
	}

	entry.Meta = &ia
	entry.Hashes = types.StringSlice(hashes)

	entry.Ids = types.StringArray(ids)
	return
}
