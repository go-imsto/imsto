package storage

import (
	"fmt"

	"github.com/cespare/xxhash"
	"github.com/spaolacci/murmur3"
)

// HashContent ...
func HashContent(data []byte) (uint64, string) {
	c := xxhash.Sum64(data)
	h1, h2 := murmur3.Sum128(data)
	return c, fmt.Sprintf("%16x%16x", h1, h2)
}
