package storage

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"hash/crc64"
	"wpst.me/calf/base"
)

type EntryId struct {
	id   string
	crc  uint64
	hash string
}

func NewEntryIdFromData(data []byte) (*EntryId, error) {
	c, m := HashContent(data)
	s := fmt.Sprintf("%x", c)
	id, err := base.BaseConvert(s, 16, 36)

	return &EntryId{id, c, m}, err
}

func NewEntryId(id string) (*EntryId, error) {
	hash, err := base.BaseConvert(id, 36, 16)
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

func HashContent(data []byte) (uint64, string) {
	c := crc64.Checksum(data, crc64.MakeTable(crc64.ISO))
	// return fmt.Sprintf("%x", s)
	// return fmt.Sprintf("%x", md5.Sum(data))
	return c, fmt.Sprintf("%x", md5.Sum(data))
}
