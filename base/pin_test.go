package base

import (
	"testing"

	"github.com/cespare/xxhash"
)

func TestPin(t *testing.T) {
	for _, cs := range []struct {
		name string
		n    int
	}{
		{"5B", 5},
		{"100B", 100},
		{"4KB", 4e3},
		{"20KB", 20e3},
		{"80KB", 80e3},
		{"10MB", 10e6},
	} {
		input := make([]byte, cs.n)
		for i := range input {
			input[i] = byte(i)
		}
		id := xxhash.Sum64(input)
		p := NewPin(id, EtGIF)
		// b := p.ID.Bytes()
		t.Logf("%4s id: %20d %14s \t%s", cs.name, p.ID, p.ID, p.Path())
		s := p.String()
		p2, err := ParsePin(s)
		if err != nil {
			t.Error(err)
		} else {
			t.Logf("parsed OK %q", p2)
		}

		// dt0km71q2c0rc dq0jk71n2c0oc
	}
}

func TestExt(t *testing.T) {
	for _, cs := range []struct {
		name string
		ext  ImagExt
	}{
		{"a.png", EtPNG},
		{"jpeg", EtJPEG},
	} {
		ext := ParseExt(cs.name)
		t.Logf("%8s %4s===%4s, %v", cs.name, ext, cs.ext, ext == cs.ext)
	}
}
