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
	s, err := Parse(input)

	if err != nil {
		t.Error(err)
		return
	}

	k := s.findKnot("Knot_B")
	assert.Equal(t, "This is Knot_B", k.Next().(*line).render())
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
	_, err := Parse(input)
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
	_, err := Parse(input)
	assert.NotNil(t, err)

	input = `
	-> Knot_A
	= Stitch_A
	  This is Stitch_A
	`
	_, err = Parse(input)
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
	s, err := Parse(input)
	assert.Nil(t, err)

	assert.Nil(t, s.findDivert("Unknown", nil))
	assert.Nil(t, s.findDivert("Unknown.Unknown", nil))
	assert.Nil(t, s.findDivert("Unknown.Unknown.Unknown", nil))

	assert.Equal(t, "Knot_A", s.findDivert("Knot_A", nil).(*Knot).Name())
	assert.Equal(t, s, s.findDivert("Knot_A", nil).(*Knot).Story())

	assert.Equal(t, "Stitch_A", s.findDivert("Knot_A.Stitch_A", nil).(*Stitch).Name())
	assert.Equal(t, "Stitch_B", s.findDivert("Knot_B.Stitch_B", nil).(*Stitch).Name())

	stitch := s.findDivert("Knot_A.Stitch_A", nil).(*Stitch)
	assert.Equal(t, s, stitch.Story())
	assert.Equal(t, "Stitch_A Content", stitch.Next().(*line).render())
}

func TestFindDivert(t *testing.T) {
	input := `
	-> Knot_A

	== Knot_A
		This is Knot_A ->Stitch_A
		= Stitch_A
		Stitch_A Content
			* Option A
			* Option B
			- Gather -> Stitch_B
		= Stitch_B
		Finally...
	`
	s, err := Parse(input)
	assert.Nil(t, err)

	s.Next()

	assert.Equal(t, "Knot_A", s.findDivert("Knot_A", s.Current()).(*Knot).Name())
	assert.Equal(t, 0, s.findDivertCount("Knot_A", nil))

	s.Next()
	assert.Equal(t, 1, s.findDivertCount("Knot_A", nil))

	s.Next()
	s.Next()
	s.Next()

	assert.Equal(t, "Stitch_A Content", s.Current().(*line).render())
	assert.Equal(t, "Stitch_A", s.findDivert("Stitch_A", s.Current()).(*Stitch).Name())

	s.Next()
	s.Select(0)
	s.Next()
	s.Next()
	s.Next()

	assert.Equal(t, 1, s.findDivertCount("Stitch_A", s.Current()))
}

func TestConfictKnotAndStitchName(t *testing.T) {
	input := `
	-> Knot_A

	== Knot_A
		This is Knot_A ->Stitch_A
		= Stitch_A
		Stitch_A Content
			* Option A
			* Option B
			- Gather -> Stitch_B
		= Stitch_A
		Finally...
	`
	_, err := Parse(input)
	assert.NotNil(t, err)
	assert.Equal(t, "conflict stitch name: Stitch_A", err.Error())
}
