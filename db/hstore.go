package db

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type Hstore map[string]interface{}

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

func (h *Hstore) Scan(src interface{}) (err error) {
	switch s := src.(type) {
	case string:
		*h, err = newHstore(s)
		return
	case []byte:
		*h, err = newHstore(string(s))
		return
	case map[string]interface{}:
		*h = Hstore(s)
		return
	}
	return
}

// text := `"ext"=>".jpg", "size"=>"34508", "width"=>"758", "height"=>"140", "quality"=>"93"`
func newHstore(text string) (Hstore, error) {
	re := regexp.MustCompile("\"([a-zA-Z-_]+)\"\\s?=\\>(NULL|\"([a-zA-Z0-9-_\\.]*)\"),?")
	r := strings.NewReplacer("\\\"", "\"")
	matches := re.FindAllStringSubmatch(text, -1)
	h := make(Hstore)
	for _, s := range matches {
		k, v := s[1], s[2]
		k = r.Replace(k)
		if v != "NULL" {
			v = r.Replace(s[3])
			h[k] = v
		} else {
			h[k] = nil
		}
		// fmt.Println(i, k, v)
	}

	return h, nil
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
