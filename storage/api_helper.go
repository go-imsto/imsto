package storage

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"log"
	"time"
)

type apiToken struct {
	salt  []byte
	hash  []byte
	stamp int64
	value []byte
}

const (
	min_salt_length  = 4
	min_token_length = 29
	min_life_time    = 15 * int64(time.Minute) // minutes
)

func unixStamp() int64 {
	return time.Now().UnixNano()
}

func NewToken(salt []byte) (*apiToken, error) {
	if len(salt) < min_salt_length {
		return nil, errors.New("salt is too short")
	}
	return &apiToken{salt: salt, stamp: unixStamp()}, nil
}

func (a *apiToken) VerifyString(str string) (bool, error) {
	s, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return false, err
	}
	return a.Verify(s)
}

func (a *apiToken) Verify(s []byte) (bool, error) {
	if len(s) < min_token_length {
		return false, errors.New("api: invalid token size")
	}
	log.Printf("token %x", s)

	hash := s[0:20] // sha1.Size
	log.Printf("hash  %x", hash)
	stamp := BytesToInt64(s[20:28])

	log.Printf("stamp %d", stamp)

	// timeout
	if isExpired(stamp) {
		log.Printf("token %x expired", s)
		return false, nil
	}

	value := s[28:]

	// invalid hash
	if !bytes.Equal(hash, hashToken(a.salt, value, stamp)) {
		log.Printf("invalid token %x", s)
		return false, nil
	}

	a.value = value

	return true, nil
}

func (a *apiToken) IsExpired() bool {
	return isExpired(a.stamp)
}

func (a *apiToken) GetValue() []byte {
	return a.value
}

func (a *apiToken) SetValue(value []byte) {
	a.value = value
	a.computeHash()
}

func (a *apiToken) Binary() []byte {
	in := append(a.hash, Int64ToBytes(a.stamp)...)
	log.Printf("in %x", in)
	in = append(in, a.value...)
	return in
}

func (a *apiToken) String() string {
	s := base64.StdEncoding.EncodeToString(a.Binary())
	// return fmt.Printf("%x%x%x", a.hash, a.stamp, a.value)
	return s
}

func (a *apiToken) computeHash() {
	a.stamp = unixStamp()
	a.hash = hashToken(a.salt, a.value, a.stamp)
}

func hashToken(salt, value []byte, stamp int64) []byte {
	in := append(value, Int64ToBytes(stamp)...)
	in = append(in, salt...)
	s := sha1.Sum(in)
	return s[0:20]
}

func isExpired(stamp int64) bool {
	now := unixStamp()
	log.Printf("now: %d stamp %d, interval %d", now, stamp, now-stamp)
	if now >= stamp && now-stamp < min_life_time {
		return false
	}
	return true
}

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}
