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
		assert.Equal(t, "Knot_A", c.options[4].Next().(*Knot).name)
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
	s.Next()

	assert.Equal(t, 3, len(s.Current().(*Choices).Options()))

	s.Select(0)

	s.Next()
	s.Next()
	s.Next()

	assert.Equal(t, 2, len(s.Current().(*Choices).Options()))

	// Select the first option again
	opt := s.Select(0)
	assert.Equal(t, "Opt_B", opt.Render(false))

	s.Next()
	s.Next()
	s.Next()

	// sticky
	assert.Equal(t, 2, len(s.Current().(*Choices).Options()))
}

func TestConditionalOption(t *testing.T) {
	input := `
	* {conditional_a > 0} ABC
	+ {conditional_b } DEF
	* GHI { conditional_c } JKL
	`

	s, err := parse(input)
	assert.Nil(t, err)

	s.Next()

	choices := s.Current().(*Choices)
	options := choices.options
	// options := choices.Options()

	assert.Equal(t, "conditional_a > 0", options[0].Condition().raw)
	assert.Equal(t, " ABC", options[0].Render(false))
	assert.Equal(t, "conditional_b", options[1].Condition().raw)
	assert.Equal(t, " DEF", options[1].Render(false))
	assert.Nil(t, options[2].condition)
}

func TestLabelledOption(t *testing.T) {
	input := `
	* {conditional_a > 0} ( label_a ) ABC
	+ {conditional_b } (label_b) abc[DEF]def
	* GHI { conditional_c } JKL
	`
	s, err := parse(input)
	assert.Nil(t, err)

	s.Next()
	choices := s.Current().(*Choices)
	options := choices.options
	assert.Equal(t, "label_a", options[0].path)
	assert.Equal(t, "label_b", options[1].path)

	assert.Equal(t, " abcDEF", options[1].Render(true))
}

func TestLablledOptionAndGather(t *testing.T) {
	input := `
	-> Knot_A
	== Knot_A ==
	* Option A
	* (lable_b) Option B
	- (lable_g) Gather -> Stitch_A // label will overwrite inline's path
	= Stitch_A
	* {Knot_A > 0} ( lable_a ) ABC
	+ {Knot_A__lable_g > 0} (lable_b) abc[DEF]def
	* GHI { conditional_c } JKL
	`
	s, err := parse(input)
	assert.Nil(t, err)

	s.Next()
	s.Next()
	s.Next()

	assert.Equal(t, 2, len(s.Current().(*Choices).Options()))
	assert.Equal(t, "Knot_A__lable_b", s.Current().(*Choices).Options()[1].Path())

	s.Select(0)
	s.Next()

	assert.Equal(t, "Knot_A__lable_g", s.Current().(*Gather).Path())
	s.Next()
	s.Next()

	assert.Equal(t, 1, s.objCount["Knot_A__lable_g"])
	options := s.Current().(*Choices).Options()
	assert.Equal(t, 3, len(options))

	assert.Equal(t, "Knot_A__Stitch_A__lable_a", options[0].Path())
}

func TestDuplicatedLabel(t *testing.T) {
	input := `
    * (lable_a) Option A
    * (lable_b) Option B
    + (lable_a)Option B
	`
	_, err := parse(input)
	assert.Equal(t, "duplicated label: lable_a", err.Error())

	input = `
	-> Knot_A
	== Knot_A
	This is Knot A content. -> Stitch_A
	= Stitch_A
    * (lable_a) Option A
    * (lable_b) Option B -> Knot_B
	== Knot_B
    + (lable_a)Option B
	= Stitch_A
    * (lable_a) Option A
    * (lable_b) Option B
	`
	_, err = parse(input)
	assert.Nil(t, err)
}

func TestChoicesParseError(t *testing.T) {
	input := `
    * (lable_a) Option A
    * (lable_b) Option B
    + (lable_c) --> Option B
	`
	_, err := parse(input)
	assert.NotNil(t, err)
}

func TestInvalidOptions(t *testing.T) {
	input := `
    * {'lable_a'} Option A
    * (lable_b) Option B
    + (lable_c) -> Option B
	`
	s, err := parse(input)
	assert.Nil(t, err)

	p := assert.PanicTestFunc(func() {
		s.Next().(*Choices).Options()
	})

	assert.Panics(t, p)
}
