package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JsonKV map[string]interface{}

func (m JsonKV) IsEmpty() bool {
	return len(m) == 0
}

func (m *JsonKV) Merge(other JsonKV) {
	meta := *m
	for k, v := range other {
		meta[k] = v
	}
	*m = meta
}

func (meta JsonKV) Get(key string) (v interface{}, ok bool) {
	v, ok = meta[key]
	return
}

func (meta JsonKV) Set(k string, v interface{}) {
	if meta == nil {
		meta = JsonKV{}
	}
	meta[k] = v
}

func (m JsonKV) Unset(k string) {
	delete(m, k)
}

func (meta JsonKV) Filter(keys ...string) (out JsonKV) {
	out = JsonKV{}
	for _, k := range keys {
		if v, ok := meta[k]; ok {
			out[k] = v
		}
	}
	return
}

func (m *JsonKV) Scan(b interface{}) error {
	if b == nil {
		*m = nil
		return nil
	}
	return json.Unmarshal(b.([]byte), m)
}

func (m JsonKV) Value() (driver.Value, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (z JsonKV) ToMaps(h map[string][]string) {
	h = make(map[string][]string)
	for k, v := range z {
		h[k] = []string{fmt.Sprint(v)}
	}
	return
}
