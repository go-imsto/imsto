package storage

import (
	"fmt"
	"log"
	"time"
	"wpst.me/calf/db"
	"wpst.me/calf/image"
)

func NewEntryConvert(id, name, path string, size uint32, meta, sev db.Hstore, hashes, ids []string, created time.Time) (entry *Entry, err error) {

	var eid *EntryId
	eid, err = NewEntryId(id)

	if err != nil {
		log.Println(err)
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
		Mime:    fmt.Sprint(meta.Get("mime")),
		sev:     sev,
		Created: created,
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
