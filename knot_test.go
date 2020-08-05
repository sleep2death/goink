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

func TestStitchNameConflict(t *testing.T) {
	input := `
	-> Knot_A

	== Knot_A
		This is Knot_A
		= Stitch_A
		This is Stitch_A
		= Stitch_A
		This is also Stitch_A
	`
	_, err := parse(input)
	assert.NotNil(t, err)

	input = `
	-> Knot_A
	= Stitch_A
	  This is Stitch_A
	`
	_, err = parse(input)
	assert.NotNil(t, err)
}

func TestStitchParse(t *testing.T) {
	input := `
	-> Knot_A

	== Knot_A
		This is Knot_A
		= Stitch_A
		Stitch_A Content
	== Knot_B
		This is Knot_B
		= Stitch_B
		** Option A
		** Option B
		-- Gather
	`
	s, err := parse(input)
	assert.Nil(t, err)

	assert.Nil(t, s.FindDivert("Unknown"))
	assert.Nil(t, s.FindDivert("Unknown.Unknown.Unknown"))

	assert.Equal(t, "Knot_A", s.FindDivert("Knot_A").(*Knot).Name())
	assert.Equal(t, s, s.FindDivert("Knot_A").(*Knot).Story())

	assert.Equal(t, "Stitch_A", s.FindDivert("Knot_A.Stitch_A").(*Stitch).Name())
	assert.Equal(t, "Stitch_B", s.FindDivert("Knot_B.Stitch_B").(*Stitch).Name())

	stitch := s.FindDivert("Knot_A.Stitch_A").(*Stitch)
	assert.Equal(t, s, stitch.Story())
	assert.Equal(t, "Stitch_A Content", stitch.Next().(*Inline).Render())
}
