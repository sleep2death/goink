package goink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInlineParse(t *testing.T) {
	inline := newInline("This is a -> divert")
	assert.Equal(t, inline.raw, "This is a -> divert")
}
