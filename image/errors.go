package image

import (
	"errors"
)

var (
	ErrorFormat     = errors.New("Invalid or unsupported Image Format")
	ErrOrigTooSmall = errors.New("Original Image Too Small")
)
