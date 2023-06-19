package thumbs

import (
	"io"
	"time"
)

// Item item for load and read
type Item interface {
	GetID() string
	GetName() string
	GetRoof() string
	IsOrigin() bool
	GetOrigin() string
}

// File ...
type File interface {
	io.Reader
	io.Seeker

	Name() string
	Size() int64
	Modified() time.Time
}

// LoadFunc load by key and save it into a file
type LoadFunc func(Item) error

// WalkFunc ..
type WalkFunc func(f File)

type Thumber interface {
	Thumbnail(uri string) error
}
