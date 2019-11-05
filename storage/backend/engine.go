package backend

import (
	"fmt"
	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/storage/types"
	"path"
	"strings"
	"time"
)

// Meta ...
type Meta = types.Meta

type FarmFunc func(string) (Wagoner, error)

type engine struct {
	name string
	farm FarmFunc
}

const (
	minIDLength = 8
)

// ListSpec ...
type ListSpec struct {
	Delimiter string `json:"delimiter"`
	Marker    string `json:"marker"`
	Limit     int    `json:"limit"`
	Prefix    string `json:"prefix"`
}

// ListItem ...
type ListItem struct {
	Key          string     `json:"key,omitempty" xml:"Key"`
	ETag         string     `json:"etag,omitempty" xml:"ETag"`
	Size         uint32     `json:"size,omitempty" xml:"Size"`
	LastModified *time.Time `json:"modified,omitempty" xml:"LastModified"`
}

// Key ...
type Key struct {
	Cat string
	ID  string
}

func (k *Key) String() string {
	return k.Cat + "/" + k.ID
}

// Path ...
func (k Key) Path() string {
	return path.Join(k.Cat, ID2Path(k.ID))
}

// Wagoner ...
type Wagoner interface {
	Get(k Key) ([]byte, error)
	Put(k Key, data []byte, meta Meta) (Meta, error)
	List(spec ListSpec) ([]ListItem, error)
	Exists(k Key) (bool, error)
	Delete(k Key) error
}

var engines = make(map[string]engine)

// RegisterEngine Register a Engine
func RegisterEngine(name string, farm FarmFunc) {
	if farm == nil {
		panic("imsto: Register engine is nil")
	}
	if _, dup := engines[name]; dup {
		panic("imsto: Register called twice for engine " + name)
	}
	engines[name] = engine{name, farm}
}

// FarmEngine get a intance of Wagoner by a special config name
func FarmEngine(roof string) (Wagoner, error) {
	if name := config.GetEngine(roof); name != "" {
		if engine, ok := engines[name]; ok {
			return engine.farm(roof)
		}
		return nil, fmt.Errorf("invalid engine %s of %s", name, roof)
	}

	return nil, fmt.Errorf("invalid engine of %s", roof)
}

// ID2Path ...
func ID2Path(r string) string {
	if len(r) < minIDLength || strings.Index(r, "/") > 0 { // > -1 表示有
		return r
	}
	return r[0:2] + "/" + r[2:4] + "/" + r[4:]
}
