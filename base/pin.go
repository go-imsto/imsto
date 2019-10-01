package base

import (
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

const (
	// base32Text         = "0123456789abcdefghjkmnpqrstvwxyz"
	binaryVersion byte = 1
)

// PinID ...
type PinID uint64

// Bytes ...
func (z PinID) Bytes() []byte {
	return toBytes(uint64(z))
}

// String ...
func (z PinID) String() string {
	var bInt big.Int
	return bInt.SetUint64(uint64(z)).Text(36)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (z PinID) MarshalText() ([]byte, error) {
	b := []byte(z.String())
	return b, nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (z *PinID) UnmarshalText(data []byte) (err error) {
	var id PinID
	id, err = ParseID(string(data))
	*z = id
	return
}

// ParseID ...
func ParseID(s string) (PinID, error) {
	var id uint64
	var bI big.Int
	if i, ok := bI.SetString(s, 36); ok {
		id = i.Uint64()
	} else {
		return 0, fmt.Errorf("invalid id %q", s)
	}
	return PinID(id), nil
}

// Scan implements of database/sql.Scanner
func (z *PinID) Scan(src interface{}) (err error) {
	switch s := src.(type) {
	case string:
		return z.UnmarshalText([]byte(s))
	case []byte:
		return z.UnmarshalText(s)
	}
	return fmt.Errorf("'%v' is invalid PinID", src)
}

// Value implements of database/sql/driver.Valuer
func (z PinID) Value() (driver.Value, error) {
	return z.String(), nil
}

// Pin ...
type Pin struct {
	ID  PinID
	Ext ImagExt
}

// NewPin ...
func NewPin(id uint64, ext ImagExt) Pin {
	var p = Pin{ID: PinID(id), Ext: ImagExt(ext)}
	return p
}

func (p Pin) String() string {
	return p.ID.String() + "." + p.Ext.String()
}

// Path ...
func (p Pin) Path() string {
	s := p.ID.String()
	return s[0:2] + "/" + s[2:] + "." + p.Ext.String()
}

func toBytes(id uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, id)
	return buf
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (p Pin) MarshalBinary() ([]byte, error) {
	enc := []byte{binaryVersion}
	enc = append(enc, p.ID.Bytes()...)
	return append(enc, p.Ext.Val()), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (p *Pin) UnmarshalBinary(data []byte) error {
	buf := data
	if len(buf) == 0 {
		return errors.New("Pin.UnmarshalBinary: no data")
	}

	if buf[0] != binaryVersion {
		return errors.New("Pin.UnmarshalBinary: unsupported version")
	}

	if len(buf) != /*version*/ 1+ /*id*/ 8+ /*ext*/ 1 {
		return errors.New("Pin.UnmarshalBinary: invalid length")
	}
	buf = buf[1:]
	id := int64(buf[7]) | int64(buf[6])<<8 | int64(buf[5])<<16 | int64(buf[4])<<24 |
		int64(buf[3])<<32 | int64(buf[2])<<40 | int64(buf[1])<<48 | int64(buf[0])<<56

	ext := buf[8]
	*p = Pin{}
	p.ID = PinID(uint64(id))
	p.Ext = ImagExt(ext)
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (p Pin) MarshalText() ([]byte, error) {
	b := []byte(p.String())
	return b, nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (p *Pin) UnmarshalText(data []byte) (err error) {
	*p = Pin{}
	p, err = ParsePin(string(data))
	return
}

// ParsePin ...
func ParsePin(s string) (p *Pin, err error) {
	arr := strings.Split(s, ".")
	if len(arr) < 2 {
		return nil, errors.New("invalid pin data: '" + s + "'")
	}
	var id PinID
	id, err = ParseID(arr[0])
	if err != nil {
		return
	}
	p = &Pin{}
	p.ID = id
	p.Ext = ParseExt(arr[1])
	return
}

// Scan implements of database/sql.Scanner
func (p *Pin) Scan(src interface{}) (err error) {
	switch s := src.(type) {
	case string:
		return p.UnmarshalText([]byte(s))
	case []byte:
		return p.UnmarshalText(s)
	}
	return fmt.Errorf("'%v' is invalid Pin", src)
}

// Value implements of database/sql/driver.Valuer
func (p Pin) Value() (driver.Value, error) {
	return p.String(), nil
}
