package storage

import (
	"encoding/base64"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	rd := base64.NewDecoder(base64.StdEncoding, strings.NewReader(jpegData))

	w := newHasher()

	n, err := io.Copy(w, rd)
	assert.NoError(t, err)
	assert.NotZero(t, n)
	assert.Equal(t, int(jpegSize), int(w.Len()))
	assert.Equal(t, int(jpegSize), int(n))
}

func TestHashBlob(t *testing.T) {
	data, err := base64.StdEncoding.DecodeString(jpegData)
	assert.NoError(t, err)
	assert.NotZero(t, len(data))

	c, hash := HashContent(data)
	assert.NotZero(t, c)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 32)
	t.Logf("id %d, hash %s", c, hash)
	var id uint64 = 7487654245818709005
	assert.Equal(t, id, c)
	assert.Equal(t, "709e291268aea5f67a3397679b6fd9cd", hash)
}
