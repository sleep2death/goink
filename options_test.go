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
	_, err := Parse(input)
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
	s, err := Parse(input)
	assert.Nil(t, err)

	s.Next()
	assert.Equal(t, "Hello, World", s.Current().(*line).Render())

	s.Next()
	assert.Panics(t, assert.PanicTestFunc(func() { s.Current().(*options).SetNext(nil) }))

	if choices, ok := s.Current().(*options); ok {
		// Choices can not go next, always return nil
		assert.Nil(t, s.Next())

		assert.Equal(t, 1, choices.nesting)
		assert.Equal(t, 2, len(choices.list()))
		assert.Nil(t, nil, choices.choose(3))

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
	s, err := Parse(input)
	assert.Nil(t, err)

	if c, ok := s.Next().(*options); ok {
		assert.Equal(t, "ABC.", c.opts[0].render(true))
		assert.Equal(t, "ABCDEF", c.opts[0].render(false))

		assert.Equal(t, "GHI", c.opts[1].render(true))
		assert.Equal(t, "GHIJKL", c.opts[1].render(false))

		assert.Equal(t, "", c.opts[2].render(true))
		assert.Equal(t, "MNO", c.opts[2].render(false))

		assert.Equal(t, "PQR", c.opts[3].render(true))
		assert.Equal(t, "PQR", c.opts[3].render(false))

		assert.Equal(t, "Hello", c.opts[4].render(true))
		assert.Equal(t, ", World ", c.opts[4].render(false))
		assert.Equal(t, "Knot_A", c.opts[4].Next().(*knot).name)
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
	s, err := Parse(input)
	assert.Nil(t, err)

	s.Next()
	s.Select(0)
	s.Next()
	s.Next()

	assert.Equal(t, 3, len(s.Current().(*options).list()))

	s.Select(0)

	s.Next()
	s.Next()
	s.Next()

	assert.Equal(t, 2, len(s.Current().(*options).list()))

	// Select the first option again
	opt := s.Select(0)
	assert.Equal(t, "Opt_B", opt.render(false))

	s.Next()
	s.Next()
	s.Next()

	// sticky
	assert.Equal(t, 2, len(s.Current().(*options).list()))
}

func TestConditionalOption(t *testing.T) {
	input := `
	* {conditional_a > 0} ABC
	+ {conditional_b } DEF
	* GHI { conditional_c } JKL
	`

	s, err := Parse(input)
	assert.Nil(t, err)

	s.Next()

	choices := s.Current().(*options)
	options := choices.opts
	// options := choices.List()

	assert.Equal(t, "conditional_a > 0", options[0].condition.raw)
	assert.Equal(t, " ABC", options[0].render(false))
	assert.Equal(t, "conditional_b", options[1].condition.raw)
	assert.Equal(t, " DEF", options[1].render(false))
	assert.Nil(t, options[2].condition)
}

func TestLabelledOption(t *testing.T) {
	input := `
	* {conditional_a > 0} ( label_a ) ABC
	+ {conditional_b } (label_b) abc[DEF]def
	* GHI { conditional_c } JKL
	`
	s, err := Parse(input)
	assert.Nil(t, err)

	s.Next()
	choices := s.Current().(*options)
	options := choices.opts
	assert.Equal(t, "label_a", options[0].path)
	assert.Equal(t, "label_b", options[1].path)

	assert.Equal(t, " abcDEF", options[1].render(true))
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
	s, err := Parse(input)
	assert.Nil(t, err)

	s.Next()
	s.Next()
	s.Next()

	assert.Equal(t, 2, len(s.Current().(*options).list()))
	assert.Equal(t, "Knot_A__lable_b", s.Current().(*options).list()[1].Path())

	s.Select(0)
	s.Next()

	assert.Equal(t, "Knot_A__lable_g", s.Current().(*gather).Path())
	s.Next()
	s.Next()

	assert.Equal(t, 1, s.objCount["Knot_A__lable_g"])
	options := s.Current().(*options).list()
	assert.Equal(t, 3, len(options))

	assert.Equal(t, "Knot_A__Stitch_A__lable_a", options[0].Path())
}

func TestDuplicatedLabel(t *testing.T) {
	input := `
    * (lable_a) Option A
    * (lable_b) Option B
    + (lable_a)Option B
	`
	_, err := Parse(input)
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
	_, err = Parse(input)
	assert.Nil(t, err)
}

func TestChoicesParseError(t *testing.T) {
	input := `
    * (lable_a) Option A
    * (lable_b) Option B
    + (lable_c) --> Option B
	`
	_, err := Parse(input)
	assert.NotNil(t, err)
}

func TestInvalidoptions(t *testing.T) {
	input := `
    * {'lable_a'} Option A
    * (lable_b) Option B
    + (lable_c) -> Option B
	`
	s, err := Parse(input)
	assert.Nil(t, err)

	p := assert.PanicTestFunc(func() {
		s.Next().(*options).list()
	})

	assert.Panics(t, p)
}
