package db

import (
	"database/sql/driver"
	"fmt"
	// "log"
	"reflect"
	"strings"
)

type Qarray []interface{}

func NewQarrayText(s string) (Qarray, error) {
	if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		s = strings.Trim(s, "{}")
	}
	if len(s) == 0 {
		return Qarray{}, nil
	}
	a := strings.Split(s, ",")
	return NewQarray(a)

	// log.Println("invalid Qarray format")
	// return Qarray{}, nil
}

func NewQarray(a []string) (Qarray, error) {
	q := make(Qarray, len(a))
	for i := 0; i < len(a); i++ {
		q[i] = a[i]
	}
	return q, nil
}

func (q Qarray) String() string {
	if len(q) == 0 {
		return "{}"
	}

	var s = make([]string, len(q))
	r := strings.NewReplacer("\\", "\\\\", "'", "''", "\"", "\\\"")

	for i, v := range q {
		sv := reflect.ValueOf(v)
		switch sv.Kind() {
		default:
			s[i] = "\"" + r.Replace(fmt.Sprintf("%s", v)) + "\""
			break
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			s[i] = fmt.Sprintf("%v", v)
			break
		}
	}
	return "{" + strings.Join(s, ",") + "}"
}

// driver.Valuer for sql value save
func (q Qarray) Value() (driver.Value, error) {
	return q.String(), nil
}

// driver.Scanner for sql value load
func (q *Qarray) Scan(src interface{}) (err error) {
	switch s := src.(type) {
	case string:
		*q, err = NewQarrayText(s)
		return
	case []byte:
		*q, err = NewQarrayText(string(s))
		return
	case []interface{}:
		*q = Qarray(s)
		return
	}
	return
}

// Index returns the index of the string s in Qarray, or -1 if s is not found.
func (q Qarray) Index(s string) int {
	for i, v := range q {
		if vs := v.(string); s == vs {
			return i
		}
		if s == fmt.Sprint(v) {
			return i
		}
	}
	return -1
}

// Contains returns true if the string s is found
func (q Qarray) Contains(s string) bool {
	return q.Index(s) > -1
}

func (q Qarray) ToStringSlice() []string {
	a := make([]string, len(q))
	for i, v := range q {
		a[i] = fmt.Sprint(v)
	}
	return a
}
