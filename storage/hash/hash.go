package hash

import (
	"fmt"
	"io"

	"github.com/spaolacci/murmur3"
)

// Hasher ...
type Hasher interface {
	io.Writer
	Bytes() []byte
	Len() uint32
	String() string
}

type hasher struct {
	mm3 murmur3.Hash128
	n   int
}

// New ...
func New() Hasher {
	return &hasher{mm3: murmur3.New128()}
}

func (h *hasher) Write(b []byte) (n int, err error) {
	n, err = h.mm3.Write(b)
	if err != nil {
		return
	}
	h.n += len(b)
	return
}

func (h *hasher) Bytes() []byte {
	h1, h2 := h.mm3.Sum128()
	return combine(h1, h2, h.n)
}

func (h *hasher) String() string {
	return fmt.Sprintf("%x", h.Bytes())
}

func (h *hasher) Len() uint32 {
	return uint32(h.n)
}

// SumContent ...
func SumContent(data []byte) string {
	h1, h2 := murmur3.Sum128(data)
	return fmt.Sprintf("%x", combine(h1, h2, len(data)))
}

func combine(h1, h2 uint64, t int) []byte {
	return []byte{
		byte(h1 >> 56), byte(h1 >> 48), byte(h1 >> 40), byte(h1 >> 32),
		byte(h1 >> 24), byte(h1 >> 16), byte(h1 >> 8), byte(h1),

		byte(h2 >> 56), byte(h2 >> 48), byte(h2 >> 40), byte(h2 >> 32),
		byte(h2 >> 24), byte(h2 >> 16), byte(h2 >> 8), byte(h2),

		byte(t >> 8), byte(t),
	}
}
