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

type apiVer uint8
type valueCate uint8

type apiToken struct {
	ver   apiVer
	appid AppId
	salt  []byte
	hash  []byte
	stamp int64
	vc    valueCate
	value []byte
}

const (
	VC_TOKEN         valueCate = 1
	VC_TICKET        valueCate = 2
	min_salt_length            = 4
	min_token_length           = 29
	min_life_time              = 55 * int64(time.Minute) // minutes
)

const (
	di_ver   = 0
	di_app   = di_ver + 1
	di_vc    = di_app + 1
	di_hash  = di_vc + 1
	di_stamp = di_hash + 20
	di_value = di_stamp + 8
)

func unixStamp() int64 {
	return time.Now().UnixNano()
}

func newToken(ver apiVer, appid AppId, salt []byte) (*apiToken, error) {
	if len(salt) < min_salt_length {
		return nil, errors.New("salt is too short")
	}
	return &apiToken{ver: ver, appid: appid, salt: salt, stamp: unixStamp()}, nil
}

func (a *apiToken) VerifyString(str string) (bool, error) {
	s, err := base64.URLEncoding.DecodeString(str)
	if err != nil {
		log.Printf("verifystring '%s' error: %s", str, err)
		return false, err
	}
	return a.Verify(s)
}

func (a *apiToken) Verify(s []byte) (bool, error) {
	if len(s) < min_token_length {
		return false, errors.New("api: invalid token size")
	}
	log.Printf("token %x", s)
	vc := s[di_vc:di_hash]

	hash := s[di_hash:di_stamp] // sha1.Size
	log.Printf("hash  %x", hash)
	stamp := BytesToInt64(s[di_stamp:di_value])

	log.Printf("stamp %d", stamp)

	// timeout
	if isExpired(stamp) {
		log.Printf("token %x expired", s)
		return false, errors.New("api: the token is expired")
	}
	value := s[di_value:]

	// invalid hash
	if !bytes.Equal(hash, hashToken(a.salt, value, stamp)) {
		// log.Printf("invalid token %x", s)
		return false, errors.New("api: invalid token")
	}

	log.Printf("api: verify ok, and got value: %x", value)

	a.vc = valueCate(vc[0])
	a.value = value

	return true, nil
}

func (a *apiToken) IsExpired() bool {
	return isExpired(a.stamp)
}

func (a *apiToken) GetValue() []byte {
	return a.value
}

func (a *apiToken) GetValuleString() string {
	return string(a.value)
}

func (a *apiToken) GetValuleInt() int64 {
	return BytesToInt64(a.value)
}

func (a *apiToken) SetValueInt(val int, vc valueCate) {
	a.SetValue(Int64ToBytes(int64(val)), vc)
}

func (a *apiToken) SetValue(value []byte, vc valueCate) {
	a.value = value
	a.vc = vc
	a.computeHash()
}

func (a *apiToken) Binary() []byte {
	in := []byte{byte(a.ver), byte(a.appid), byte(a.vc)}
	in = append(in, a.hash...)
	in = append(in, Int64ToBytes(a.stamp)...)
	log.Printf("in %x", in)
	in = append(in, a.value...)
	return in
}

func (a *apiToken) String() string {
	s := base64.URLEncoding.EncodeToString(a.Binary())
	// return fmt.Printf("%x%x%x", a.hash, a.stamp, a.value)
	return s
}

func (a *apiToken) computeHash() {
	a.stamp = unixStamp()
	// TODO: add ver and appid to hash
	a.hash = hashToken(a.salt, a.value, a.stamp)
}

func hashToken(salt, value []byte, stamp int64) []byte {
	in := append(value, Int64ToBytes(stamp)...)
	in = append(in, salt...)
	s := sha1.Sum(in)
	return s[0:sha1.Size]
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
