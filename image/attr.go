package image

import (
	"database/sql/driver"
	"encoding/json"
)

type Dimension uint32
type Quality uint8

// Attr ...
type Attr struct {
	Width   Dimension `json:"width"`
	Height  Dimension `json:"height"`
	Quality Quality   `json:"quality,omitempty"`
	Ext     string    `json:"ext,omitempty"`
	Mime    string    `json:"mime,omitempty"`
	Name    string    `json:"name,omitempty"`
}

// ToMap ...
func (a Attr) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"width":  a.Width,
		"height": a.Height,
		"ext":    a.Ext,
		"mime":   a.Mime,
	}

	if len(a.Name) > 0 {
		m["name"] = a.Name
	}
	if a.Quality > 0 {
		m["quality"] = a.Quality
	}
	return m
}

func (a *Attr) FromMap(m map[string]interface{}) {
	if m == nil {
		return
	}
	if a == nil {
		*a = Attr{}
	}
	if v, ok := m["width"]; ok {
		if vv, ok := v.(uint32); ok {
			a.Width = Dimension(vv)
		}
	}
	if v, ok := m["height"]; ok {
		if vv, ok := v.(uint32); ok {
			a.Height = Dimension(vv)
		}
	}
	if v, ok := m["ext"]; ok {
		if vv, ok := v.(string); ok {
			a.Ext = vv
		}
	}
	if v, ok := m["mime"]; ok {
		if vv, ok := v.(string); ok {
			a.Mime = vv
		}
	}
}

func (a *Attr) Scan(b interface{}) error {
	if b == nil {
		return nil
	}
	return json.Unmarshal(b.([]byte), a)
}

func (a Attr) Value() (driver.Value, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// NewAttr ...
func NewAttr(w, h uint, ext string) *Attr {
	a := &Attr{
		Width:  Dimension(w),
		Height: Dimension(h),
		Ext:    getExt(ext),
	}
	return a
}

func getExt(f string) string {
	if f == "jpeg" {
		return ".jpg"
	}
	return "." + f
}
