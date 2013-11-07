package engine

import (
	"calf/config"
	"calf/storage"
	"errors"
)

type engine struct {
	name string
	farm func(string) (EntryMapper, error)
}

type EntryMapper interface {
	Get(key string) ([]byte, error)
	Put(entry storage.Entry, data []byte) error
	Exists(key string) bool
	Del(key string) error
}

var engines = make(map[string]engine)

func Register(name string, farm func(string) (EntryMapper, error)) {
	if farm == nil {
		panic("imsto: Register engine is nil")
	}
	if _, dup := engines[name]; dup {
		panic("imsto: Register called twice for engine " + name)
	}
	engines[name] = engine{name, farm}
}

func Farm(sn string) (em EntryMapper, err error) {
	name := config.GetValue(sn, "engine")

	if engine, ok := engines[name]; ok {
		em, err = engine.farm(sn)
		return
	}

	return nil, errors.New("invalid engine name: " + name)
}
