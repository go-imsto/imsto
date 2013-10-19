package imsto

import (
	"bytes"
)

const (
	TYPE_NONE = 0
	TYPE_GIF  = 1
	TYPE_JPEG = 2
	TYPE_PNG  = 3
)

const (
	SIG_GIF = "GIF"
	SIG_JPG = "\xff\xd8\xff"
	// SIG_PNG = "\x89\x50\x4e\x47\x0d\x0a\x1a\x0a"
	SIG_PNG = "\211PNG\r\n\032\n"
)

func GuessImageType(data *[]byte) int {
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

func ExtByImageType(it int) string {
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
