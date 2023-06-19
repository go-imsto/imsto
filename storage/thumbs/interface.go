package thumbs

import (
	"io"
	"time"
)

// File ...
type File interface {
	io.Reader
	io.Seeker

	Name() string
	Size() int64
	Modified() time.Time
}

// LoadFunc load by key and save it into a file
type LoadFunc func(key string, orig string) error

// WalkFunc ..
type WalkFunc func(f File)

type Thumber interface {
	Thumbnail(uri string) error
}
