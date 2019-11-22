package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
TestConfig ...

env:

IMSTO_ENGINES='demo:file'
IMSTO_PREFIXES='demo:demos'


*/
func TestConfig(t *testing.T) {

	assert.Equal(t, "file", GetEngine("demo"))
	assert.Equal(t, "demos", GetPrefix("demo"))
}
