package storage

import (
	"crypto/sha1"
	"database/sql/driver"
	"fmt"

	"github.com/cespare/xxhash"

	"github.com/go-imsto/imsto/base"
)

type EntryId struct {
	base.Pin
	hash string
}

// const (
// 	BASE_SRC = 16
// 	BASE_DST = 36
// )

func NewEntryIdFromData(data []byte, name string) (*EntryId, error) {
	id, hash := HashContent(data)
	pin := base.NewPin(id, base.ParseExt(name))

	return &EntryId{pin, hash}, nil
}

func NewEntryId(s string) (*EntryId, error) {
	id, err := base.ParseID(s)
	if err != nil {
		return nil, err
	}
	pin := base.Pin{ID: id}
	return &EntryId{Pin: pin}, nil
}

func (ei *EntryId) String() string {
	return ei.Pin.String()
}

func (ei *EntryId) MarshalText() ([]byte, error) {
	return ei.Pin.MarshalText()
}

func (ei *EntryId) Hashed() string {
	return ei.ID.String()
}

func (ei *EntryId) tip() string {
	return ei.Pin.ID.String()[:1]
}

func (ei *EntryId) Scan(src interface{}) (err error) {
	switch s := src.(type) {
	case string:
		ei, err = NewEntryId(s)
		return
	case []byte:
		ei, err = NewEntryId(string(s))
		return
	}
	return fmt.Errorf("'%s' is invalid entryId", src)
}

func (ei EntryId) Value() (driver.Value, error) {
	return ei.ID.String(), nil
}

func HashContent(data []byte) (uint64, string) {
	c := xxhash.Sum64(data)
	return c, fmt.Sprintf("%x", sha1.Sum(data))
}
