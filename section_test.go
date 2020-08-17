package goink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoryGoOn(t *testing.T) {
	input := `
	Hello World,
	* [Option] A #tag a
	* Option B
	- -> Knot_A
	== Knot_A
	A content here.
	`
	s, err := Parse(input)
	assert.Nil(t, err)

	state := NewState(s, true)

	sec, err := s.GoOn(state)
	assert.Nil(t, err)

	state = s.Save()

	assert.Equal(t, "\nHello World,", sec.text)
	assert.Equal(t, 2, len(sec.opts))
	assert.Equal(t, "tag a", sec.optsTags[0][0])

	sec, err = s.Select(state, 0)
	assert.Nil(t, err)
	assert.Equal(t, "\n A \nA content here.", sec.text)
}
