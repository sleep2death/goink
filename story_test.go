package goink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadLines(t *testing.T) {
	lines, err := readLines("./ink/sample.ink")
	assert.Nil(t, err)
	t.Log(lines)
	// assert.Equal(t, 9, len(lines))
}
