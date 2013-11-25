package storage

import (
	"errors"
	"wpst.me/calf/config"
	"wpst.me/calf/db"
)

type engine struct {
	name string
	farm func(string) (Wagoner, error)
}

type Wagoner interface {
	Get(key string) ([]byte, error)
	Put(key string, data []byte, meta db.Hstore) (db.Hstore, error)
	Exists(key string) (bool, error)
	Del(key string) error
}

var engines = make(map[string]engine)

// Register a Engine
func RegisterEngine(name string, farm func(string) (Wagoner, error)) {
	if farm == nil {
		panic("imsto: Register engine is nil")
	}
	if _, dup := engines[name]; dup {
		panic("imsto: Register called twice for engine " + name)
	}
	engines[name] = engine{name, farm}
}

// get a intance of Wagoner by a special engine name
func FarmEngine(sn string) (em Wagoner, err error) {
	name := config.GetValue(sn, "engine")

	if engine, ok := engines[name]; ok {
		em, err = engine.farm(sn)
		return
	}

	return nil, errors.New("invalid engine name: " + name)
}
