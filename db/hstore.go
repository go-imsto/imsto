package db

import (
	"database/sql/driver"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"regexp"
	"strings"
)

const hstore_pattern = "\"([a-zA-Z-_]+)\"\\s?=\\>(NULL|\"([a-zA-Z0-9-_\\.]*)\"),?"

type Hstore map[string]interface{}

// text := `"ext"=>".jpg", "size"=>"34508", "width"=>"758", "height"=>"140", "quality"=>"93"`
func NewHstore(text string) (Hstore, error) {
	h := make(Hstore)
	err := h.fill(text)
	if err != nil {
		return nil, err
	}
	return h, nil
}

// hstore map 转成 string 值
func (h Hstore) String() string {
	var a = make([]string, len(h))
	r := strings.NewReplacer("\\", "\\\\", "'", "''", "\"", "\\\"")
	i := 0
	for k, v := range h {
		k = r.Replace(k)
		if v == nil {
			a[i] = "\"" + k + "\"" + "=>" + "NULL"
		} else {
			a[i] = fmt.Sprintf("\"%s\"=>\"%s\"", k, r.Replace(fmt.Sprint(v)))
		}
		i++
	}

	return strings.Join(a, ",")
}

// driver.Valuer for sql value save
func (h Hstore) Value() (driver.Value, error) {
	return h.String(), nil
}

// driver.Scanner for sql value load
func (h *Hstore) Scan(src interface{}) (err error) {
	switch s := src.(type) {
	case string:
		*h, err = NewHstore(s)
		return
	case []byte:
		*h, err = NewHstore(string(s))
		return
	case map[string]interface{}:
		*h = Hstore(s)
		return
	}
	return
}

func (h Hstore) Get(k string) (v interface{}) {
	v = h[k]
	return
}

func (h Hstore) Set(k string, v interface{}) {
	h[k] = v
}

func (h *Hstore) fill(text string) error {
	re, err := regexp.Compile(hstore_pattern)
	if err != nil {
		return err
	}
	r := strings.NewReplacer("\\\"", "\"")
	matches := re.FindAllStringSubmatch(text, -1)
	for _, s := range matches {
		k, v := s[1], s[2]
		k = r.Replace(k)
		if v != "NULL" {
			v = r.Replace(s[3])
			// h[k] = v
			h.Set(k, v)
		} else {
			// h[k] = nil
			h.Set(k, nil)
		}
		// fmt.Println(i, k, v)
	}
	return nil
}

func (h Hstore) ToStruct(i interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		WeaklyTypedInput: true,
		Result:           i,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(h)
}

type Hstorer interface {
	Hstore() Hstore
}

func StructToHstore(i interface{}) Hstore {
	h := make(Hstore)
	iVal := reflect.ValueOf(i)
	typ := iVal.Type()
	for i := 0; i < iVal.NumField(); i++ {
		f := iVal.Field(i)
		k := strings.ToLower(typ.Field(i).Name)
		switch v := f.Interface().(type) {
		default:
			h[k] = v
		}
	}
	return h
}

type NullHstore struct {
	Hstore Hstore
	Valid  bool // Valid is true if Hstore is not NULL
}

// Scan implements the Scanner interface.
func (n *NullHstore) Scan(value interface{}) error {
	n.Hstore, n.Valid = value.(Hstore)
	return nil
}

// Value implements the driver Valuer interface.
func (n NullHstore) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Hstore, nil
}
