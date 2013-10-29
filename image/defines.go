package image

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
)

const (
	TYPE_NONE = 0
	TYPE_GIF  = 1
	TYPE_JPEG = 2
	TYPE_PNG  = 3
)

const (
	SIG_GIF = "GIF8"
	SIG_JPG = "\xff\xd8\xff"
	// SIG_PNG = "\x89\x50\x4e\x47\x0d\x0a\x1a\x0a"
	SIG_PNG = "\211PNG\r\n\032\n"
)

func GuessType(data *[]byte) int {
	if bytes.HasPrefix(*data, []byte(SIG_GIF)) {
		return TYPE_GIF
	}

	if bytes.HasPrefix(*data, []byte(SIG_JPG)) {
		return TYPE_JPEG
	}

	if bytes.HasPrefix(*data, []byte(SIG_PNG)) {
		return TYPE_PNG
	}

	return TYPE_NONE
}

func ExtByType(it int) string {
	switch it {
	case TYPE_GIF:
		return ".gif"
	case TYPE_JPEG:
		return ".jpg"
	case TYPE_PNG:
		return ".png"
	default:
		return ""
	}
}

var (
	ErrorFormat = errors.New("Invalid or unsupported Image Format")
)

const _head_size = 8

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

func readHeadFile(filename string) ([]byte, error) {

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := bufio.NewReaderSize(file, _head_size)
	// fmt.Println(data)
	return readHead(r)
}

func readHead(r io.Reader) ([]byte, error) {
	rr := asReader(r)
	return rr.Peek(_head_size)
}
