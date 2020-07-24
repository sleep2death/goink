package goink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChoiceParse(t *testing.T) {
	s := &Story{}
	c, err := s.parse("* Choice 1")

	assert.Nil(t, err)
	assert.Equal(t, 1, c.(*Choice).nesting)
	assert.Equal(t, CStar, c.(*Choice).ct)

	cc, err := c.parse("++ Choice 1.1")
	assert.Nil(t, err)
	assert.Equal(t, c, cc.Parent())
	assert.Equal(t, CPlus, cc.(*Choice).ct)

	ccc, err := cc.parse("*** Choice 1.1.1")
	assert.Nil(t, err)
	assert.Equal(t, cc, ccc.Parent())

	ccc, err = ccc.parse("*** Choice 1.1.2")
	assert.Nil(t, err)
	assert.Equal(t, cc, ccc.Parent())

	c, err = ccc.parse("* Choice 2")
	assert.Nil(t, err)
	assert.Equal(t, s, c.Parent())

	k, err := c.parse("== Knot_1")
	assert.Nil(t, err)
	assert.Equal(t, s, k.Parent())

	c, err = k.parse("** Choice 3")
	assert.NotNil(t, err)

	c, err = k.parse("* Choice 3")
	assert.Equal(t, k, c.Parent())

	c, err = k.parse("=== Knot_1")
	assert.NotNil(t, err)
}
