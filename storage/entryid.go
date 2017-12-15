package storage

import (
	"crypto/md5"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	// "hash/crc64"
	"github.com/OneOfOne/xxhash"
	"wpst.me/calf/base"
)

type EntryId struct {
	id   string
	crc  uint64
	hash string
}

const (
	BASE_SRC = 16
	BASE_DST = 36
)

func NewEntryIdFromData(data []byte) (*EntryId, error) {
	c, m := HashContent(data)
	s := fmt.Sprintf("%x", c)
	id, err := base.Convert(s, BASE_SRC, BASE_DST)

	return &EntryId{id, c, m}, err
}

func NewEntryId(id string) (*EntryId, error) {
	hash, err := base.Convert(id, BASE_DST, BASE_SRC)
	return &EntryId{id, 0, hash}, err
}

func (ei *EntryId) String() string {
	return ei.id
}

func (ei *EntryId) MarshalJSON() ([]byte, error) {
	return json.Marshal(ei.id)
}

func (ei *EntryId) Hashed() string {
	return ei.hash
}

func (ei *EntryId) tip() string {
	return ei.id[:1]
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
	return ei.id, nil
}

func HashContent(data []byte) (uint64, string) {
	// c := crc64.Checksum(data, crc64.MakeTable(crc64.ISO))
	c := xxhash.Checksum64(data)
	// return fmt.Sprintf("%x", s)
	// return fmt.Sprintf("%x", md5.Sum(data))
	return c, fmt.Sprintf("%x", md5.Sum(data))
}
