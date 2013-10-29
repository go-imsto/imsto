package image

import (
	"bufio"
	// "bytes"
	"errors"
	"io"
	// "os"
	"log"
)

type TypeId int

const (
	TYPE_NONE TypeId = iota
	TYPE_GIF
	TYPE_JPEG
	TYPE_PNG
)

var type_labels = [...]string{
	"None",
	"GIF",
	"JPEG",
	"PNG",
}

func (id TypeId) String() string { return type_labels[id] }

type format struct {
	id        TypeId
	sign, ext string
}

// "\211PNG\r\n\032\n"
var formats = []format{
	format{TYPE_GIF, "GIF8?a", ".gif"},
	format{TYPE_JPEG, "\xff\xd8\xff", ".jpg"},
	format{TYPE_PNG, "\x89PNG\r\n\x1a\n", ".jpg"}, // SIG_PNG = "\x89\x50\x4e\x47\x0d\x0a\x1a\x0a"
}

func registerFormat(t TypeId, sign, ext string) {
	formats = append(formats, format{t, sign, ext})
}

func match(sign string, b []byte) bool {
	if len(sign) != len(b) {
		return false
	}
	for i, c := range b {
		if sign[i] != c && sign[i] != '?' {
			return false
		}
	}
	return true
}

func sniff(r reader) format {
	for _, f := range formats {
		b, err := r.Peek(len(f.sign))
		if err != nil {
			log.Println(err)
		}
		if err == nil && match(f.sign, b) {
			return f
		}
	}
	return format{id: TYPE_NONE}
}

func GuessType(r io.Reader) (TypeId, string, error) {
	rr := asReader(r)
	f := sniff(rr)

	if f.id == TYPE_NONE {
		return TYPE_NONE, "", ErrorFormat
	}

	return f.id, f.ext, nil
}

var (
	ErrorFormat = errors.New("Invalid or unsupported Image Format")
)

// A reader is an io.Reader that can also peek ahead.
type reader interface {
	io.Reader
	Peek(int) ([]byte, error)
}

// asReader converts an io.Reader to a reader.
func asReader(r io.Reader) reader {
	if rr, ok := r.(reader); ok {
		return rr
	}
	return bufio.NewReader(r)
}
