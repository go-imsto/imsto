package storage

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
)

type hstore map[string]interface{}

func (h hstore) String() string {
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

// text := `"ext"=>".jpg", "size"=>"34508", "width"=>"758", "height"=>"140", "quality"=>"93"`
func newHstore(text string) (hstore, error) {
	re := regexp.MustCompile("\"([a-zA-Z-_]+)\"\\s?=\\>(NULL|\"([a-zA-Z0-9-_\\.]*)\"),?")
	r := strings.NewReplacer("\\\"", "\"")
	matches := re.FindAllStringSubmatch(text, -1)
	h := make(hstore)
	for i, s := range matches {
		k, v := s[1], s[2]
		k = r.Replace(k)
		if v != "NULL" {
			v = r.Replace(s[3])
			h[k] = v
		} else {
			h[k] = nil
		}
		fmt.Println(i, k, v)
	}

	return h, nil
}

type NullHstore struct {
	Hstore hstore
	Valid  bool // Valid is true if Hstore is not NULL
}

// Scan implements the Scanner interface.
func (n *NullHstore) Scan(value interface{}) error {
	n.Hstore, n.Valid = value.(hstore)
	return nil
}

// Value implements the driver Valuer interface.
func (n NullHstore) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Hstore, nil
}
