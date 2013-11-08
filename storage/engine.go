package storage

import (
	"calf/config"
	"calf/db"
	"errors"
)

type engine struct {
	name string
	farm func(string) (Wagoner, error)
}

type Wagoner interface {
	Get(key string) ([]byte, error)
	Put(key string, data []byte, mime string) (db.Hstore, error)
	Exists(key string) bool
	Del(key string) error
}

var engines = make(map[string]engine)

func RegisterEngine(name string, farm func(string) (Wagoner, error)) {
	if farm == nil {
		panic("imsto: Register engine is nil")
	}
	if _, dup := engines[name]; dup {
		panic("imsto: Register called twice for engine " + name)
	}
	engines[name] = engine{name, farm}
}

func FarmEngine(sn string) (em Wagoner, err error) {
	name := config.GetValue(sn, "engine")

	if engine, ok := engines[name]; ok {
		em, err = engine.farm(sn)
		return
	}

	return nil, errors.New("invalid engine name: " + name)
}
