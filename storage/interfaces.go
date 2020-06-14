package storage

import (
	cdb "github.com/go-imsto/imsto/storage/types"
)

// MetaFilter ...
type MetaFilter struct {
	Tags   string
	App    AppID
	Author Author
}

// MetaWrapper ...
type MetaWrapper interface {
	Browse(limit, offset int, sort map[string]int, filter MetaFilter) ([]*Entry, error)
	Count(filter MetaFilter) (int, error)
	NextID() (uint64, error)
	Ready(entry *Entry) error
	SetDone(id string, sev cdb.Meta) error
	Save(entry *Entry, isUpdate bool) error
	BatchSave(entries []*Entry) error
	GetMeta(id string) (*Entry, error)
	GetHash(hash string) (*HashEntry, error)
	GetMapping(id string) (*mapItem, error)
	Delete(id string) error
	MapTags(id string, tags string) error
	UnmapTags(id string, tags string) error
}

type rowScanner interface {
	Scan(...interface{}) error
}
