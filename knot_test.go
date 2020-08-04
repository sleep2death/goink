package goink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKnotParse(t *testing.T) {
	input := `
	-> Knot_A

	== Knot_A ==
		This is Knot_A
		* Option A
		- Gather A
	=== Knot_B ===
		This is Knot_B
	`
	s, err := parse(input)

	if err != nil {
		t.Error(err)
		return
	}

	k := s.FindKnot("Knot_B")
	assert.Equal(t, "This is Knot_B", k.Next().(*Inline).Render())
}

func TestKnotNameConflict(t *testing.T) {
	input := `
	-> Knot_A

	== Knot_A
		This is Knot_A
	== Knot_B
		This is Knot_B
	=== Knot_A ===
		This is also Knot_A
	`
	_, err := parse(input)
	assert.NotNil(t, err)
}

func TestStitchParse(t *testing.T) {
	input := `
	-> Knot_A

	== Knot_A
		This is Knot_A
		= Stitch_A
	== Knot_B
		This is Knot_B
		= Stitch_B
		** Option A
		** Option B
	`
	s, err := parse(input)
	assert.Nil(t, err)
	assert.Equal(t, "Stitch_B", s.FindDivert(s.current, "Knot_B.Stitch_B").(*Stitch).name)
}
