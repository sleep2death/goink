package goink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadLines(t *testing.T) {
	s, err := readInk("./ink/sample.ink")
	assert.Nil(t, err)
	t.Log(len(s.content()))
	// assert.Equal(t, 9, len(lines))
}
