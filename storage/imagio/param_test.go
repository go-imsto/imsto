package imagio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParam(t *testing.T) {
	uri := "/show/s120/bdouymx4a7ro.jpg"
	p, err := ParseFromPath(uri)
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "bdouymx4a7ro.jpg", p.Name)
	assert.Equal(t, "bd/ou/ymx4a7ro.jpg", p.Path)
	assert.Equal(t, 120, int(p.Width))
	assert.Equal(t, 120, int(p.Height))
}

func TestParseSize(t *testing.T) {
	s := "s256x128"

	m, w, h, err := ParseSize(s)
	assert.NoError(t, err)
	assert.Equal(t, "s", m)
	assert.Equal(t, 256, int(w))
	assert.Equal(t, 128, int(h))
}
