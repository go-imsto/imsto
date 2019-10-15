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
	assert.Equal(t, 120, p.Width)
	assert.Equal(t, 120, p.Height)
}
