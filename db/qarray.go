package db

import (
	"database/sql/driver"
	"fmt"
	"log"
	"reflect"
	"strings"
)

type Qarray []interface{}

func NewQarrayText(text string) (Qarray, error) {
	if strings.HasPrefix(text, "{") && strings.HasSuffix(text, "}") {
		s := strings.Trim(text, "{}")
		a := strings.Split(s, ",")
		return NewQarray(a)
	}

	log.Println("invalid Qarray format")
	return Qarray{}, nil
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
