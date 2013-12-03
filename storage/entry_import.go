package storage

import (
	"errors"
	"log"
	"time"
	"wpst.me/calf/db"
	"wpst.me/calf/image"
)

func NewEntryConvert(id, name, path, mime string, size uint32, meta, sev db.Hstore, hashes, ids []string, created time.Time) (entry *Entry, err error) {

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
	err = meta.ToStruct(&ia)
	if err != nil {
		log.Println(err)
		return
	}
	// log.Println("ia:", ia)

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
	entry.Hashes, err = db.NewQarray(hashes)
	if err != nil {
		log.Println(err)
		return
	}

	entry.Ids, err = db.NewQarray(ids)
	if err != nil {
		log.Println(err)
		return
	}
	return
}
