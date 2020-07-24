package goink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewKnot(t *testing.T) {
	s := &Story{}

	// Add a knot
	k, err := s.parse("=== Knot_1 ===")
	assert.Nil(t, err)

	assert.Equal(t, "Knot_1", k.(*Knot).name)
	assert.Equal(t, 1, len(s.Content()))
	assert.Equal(t, s, k.Parent())

	assert.Nil(t, newKnot("= NotAKnot"))
	assert.Nil(t, newKnot("-- NotAKnot"))

	// Add another knot
	k, err = s.parse("=== Knot_2 ===")
	assert.Equal(t, 2, len(s.Content()))

	// Add a duplicated knot, should return error
	k, err = s.parse("=== Knot_1 ===")
	assert.NotNil(t, err)
	assert.Equal(t, 2, len(s.Content()))
}

func TestKnotParse(t *testing.T) {
	s := &Story{}

	// Add a knot
	k, err := s.parse("=== Knot_1 ===")
	assert.Nil(t, err)

	k, err = k.parse("=== Knot_2 ===")
	assert.Equal(t, 2, len(s.Content()))

	// Add an invalid choice to knot
	c, err := k.parse("** Choice_1")
	assert.NotNil(t, err)

	// Add a choice to knot
	c, err = k.parse("* Choice_1")
	assert.Equal(t, k, c.Parent())

	// Add an knot
	c, err = k.parse("* Choice_1")
	assert.Nil(t, err)

	k, err = k.parse("=== Knot_1 ===")
	assert.NotNil(t, err)
}
