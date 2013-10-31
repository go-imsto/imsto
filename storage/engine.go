package storage

type EntryMapper interface {
	Get(key string) (Entry, error)
	Put(key string, data []byte) error
	Exists(key string)
}

var engines = make(map[string]EntryMapper)

func Register(name string, engine EntryMapper) {
	if engine == nil {
		panic("imsto: Register engine is nil")
	}
	if _, dup := engines[name]; dup {
		panic("imsto: Register called twice for engine " + name)
	}
	engines[name] = engine
}

func Open(section string) {
	// TODO:
}
