package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Meta JSON KV for sql
type Meta map[string]interface{}

// IsEmpty ...
func (m Meta) IsEmpty() bool {
	return len(m) == 0
}

// Merge ...
func (m *Meta) Merge(other Meta) {
	meta := *m
	for k, v := range other {
		meta[k] = v
	}
	*m = meta
}

// Get ...
func (m Meta) Get(key string) (v interface{}, ok bool) {
	v, ok = m[key]
	return
}

// Set ...
func (m Meta) Set(k string, v interface{}) {
	if m == nil {
		m = Meta{}
	}
	m[k] = v
}

// Unset ...
func (m Meta) Unset(k string) {
	delete(m, k)
}

// Filter ...
func (m Meta) Filter(keys ...string) (out Meta) {
	out = Meta{}
	for _, k := range keys {
		if v, ok := m[k]; ok {
			out[k] = v
		}
	}
	return
}

// Scan ...
func (m *Meta) Scan(b interface{}) error {
	if b == nil {
		*m = nil
		return nil
	}
	return json.Unmarshal(b.([]byte), m)
}

// Value ...
func (m Meta) Value() (driver.Value, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// ToMaps ...
func (m Meta) ToMaps(h map[string][]string) {
	h = make(map[string][]string)
	for k, v := range m {
		h[k] = []string{fmt.Sprint(v)}
	}
	return
}
