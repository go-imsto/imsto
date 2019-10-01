package base

import (
	"strings"
)

// ImagExt ...
type ImagExt byte

const (
	EtNone ImagExt = iota
	EtGIF
	EtJPEG
	EtPNG
)

func (z ImagExt) String() string {
	switch z {
	case EtGIF:
		return "gif"
	case EtJPEG:
		return "jpeg"
	case EtPNG:
		return "png"
	}
	return "unknown"
}

// Val ...
func (z ImagExt) Val() byte {
	return byte(z)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (z ImagExt) MarshalText() ([]byte, error) {
	b := []byte(z.String())
	return b, nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (z *ImagExt) UnmarshalText(data []byte) error {
	*z = ParseExt(string(data))
	return nil
}

// ParseExt ...
func ParseExt(s string) ImagExt {
	if pos := strings.LastIndex(s, "."); pos != -1 && pos < len(s) {
		s = s[pos+1:]
	}
	switch s {
	case "gif":
		return EtGIF
	case "jpeg", "jpg":
		return EtJPEG
	case "png":
		return EtPNG
	}
	return EtNone
}
