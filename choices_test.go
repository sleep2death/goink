package goink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidOptionNesting(t *testing.T) {
	input := `
	Hello World,
	* This is Option_A
		** Nesting Option_A.1
		**** Invalid Nesting Option_A.1.1
	* This is Option_B
	`
	_, err := parse(input)
	assert.NotNil(t, err)
}

func TestChoicesFunctions(t *testing.T) {
	input := `
	Hello, World

	* This is Option_A
		** Nesting Option_A.1
		** Nesting Option_A.2
		-- Gather A

	* This is Option_B

	- Final Gather
	`
	s, err := parse(input)
	assert.Nil(t, err)

	s.Next()
	assert.Equal(t, "Hello, World", s.Current().(*Inline).Render())

	s.Next()
	assert.Panics(t, assert.PanicTestFunc(func() { s.Current().(*Choices).SetNext(nil) }))

	if choices, ok := s.Current().(*Choices); ok {
		// Choices can not go next, always return nil
		assert.Nil(t, s.Next())

		assert.Equal(t, 1, choices.Nesting())
		assert.Equal(t, 2, len(choices.Options()))
		assert.Nil(t, nil, choices.Select(3))

		assert.Equal(t, s, choices.Story())
	} else {
		t.Error("should be choices type")
	}

}

func TestChoicesSupressing(t *testing.T) {
	input := `
	* ABC[.]DEF
	* GHI[]JKL
	* []MNO
	* PQR[]
	* [Hello], World -> Knot_A
	== Knot_A
	This is Knot A.
	`
	s, err := parse(input)
	assert.Nil(t, err)

	if c, ok := s.Next().(*Choices); ok {
		assert.Equal(t, "ABC.", c.options[0].Render(true))
		assert.Equal(t, "ABCDEF", c.options[0].Render(false))

		assert.Equal(t, "GHI", c.options[1].Render(true))
		assert.Equal(t, "GHIJKL", c.options[1].Render(false))

		assert.Equal(t, "", c.options[2].Render(true))
		assert.Equal(t, "MNO", c.options[2].Render(false))

		assert.Equal(t, "PQR", c.options[3].Render(true))
		assert.Equal(t, "PQR", c.options[3].Render(false))

		assert.Equal(t, "Hello", c.options[4].Render(true))
		assert.Equal(t, ", World ", c.options[4].Render(false))
		assert.Equal(t, "This is Knot A.", c.options[4].Next().(*Inline).Render())
	} else {
		t.Error("current is not choices")
	}
}

func TestStickyOption(t *testing.T) {
	input := `
	* Hello, World -> Knot_A
	== Knot_A
	* Opt_A
	+ Opt_B
	+ Opt_C
	- Loop Gather -> Knot_A
	`
	s, err := parse(input)
	assert.Nil(t, err)

	s.Next()
	s.Select(0)
	s.Next()

	assert.Equal(t, 3, len(s.Current().(*Choices).Options()))

	s.Select(0)
	s.Next()
	s.Next()

	assert.Equal(t, 2, len(s.Current().(*Choices).Options()))

	// Select the first option again
	opt := s.Select(0)
	assert.Equal(t, "Opt_B", opt.Render(false))

	s.Next()
	s.Next()

	// sticky
	assert.Equal(t, 2, len(s.Current().(*Choices).Options()))
}
