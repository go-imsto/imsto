package storage

import (
	"fmt"

	"github.com/cespare/xxhash"
	"github.com/spaolacci/murmur3"
)

type hasher struct {
	xxh *xxhash.Digest
	mm3 murmur3.Hash128
	n   int
}

func newHasher() *hasher {
	return &hasher{
		xxh: xxhash.New(),
		mm3: murmur3.New128(),
	}
}

func (h *hasher) Write(b []byte) (n int, err error) {
	n, err = h.xxh.Write(b)
	if err != nil {
		return
	}
	n, err = h.mm3.Write(b)
	if err != nil {
		return
	}
	h.n += len(b)
	return
}

func (h *hasher) Hash() (uint64, string) {
	h1, h2 := h.mm3.Sum128()
	return h.xxh.Sum64(), fmt.Sprintf("%x", combineHash(h1, h2))
}

func (h *hasher) Len() uint32 {
	return uint32(h.n)
}

// HashContent ...
func HashContent(data []byte) (uint64, string) {
	c := xxhash.Sum64(data)
	h1, h2 := murmur3.Sum128(data)
	return c, fmt.Sprintf("%x", combineHash(h1, h2))
}

func combineHash(h1, h2 uint64) []byte {
	return []byte{
		byte(h1 >> 56), byte(h1 >> 48), byte(h1 >> 40), byte(h1 >> 32),
		byte(h1 >> 24), byte(h1 >> 16), byte(h1 >> 8), byte(h1),

		byte(h2 >> 56), byte(h2 >> 48), byte(h2 >> 40), byte(h2 >> 32),
		byte(h2 >> 24), byte(h2 >> 16), byte(h2 >> 8), byte(h2),
	}
}
