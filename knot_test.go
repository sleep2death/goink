package goink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKnotParsing(t *testing.T) {
	input := `
	this is a basic parsing test
	-> Knot_A
	== Knot_A
	this is knot a -> stitch_a
	= stitch_a
	this is stitch a -> end
	`

	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	assert.Equal(t, "knot_a", story.paths["knot_a"].(*knot).Name())
	assert.Equal(t, story, story.paths["knot_a"].(*knot).Story())

	assert.Equal(t, "stitch_a", story.paths["knot_a__stitch_a"].(*stitch).Name())
	assert.Equal(t, story, story.paths["knot_a__stitch_a"].(*stitch).Story())

	ctx := NewContext()
	_, err = story.Resume(ctx)
	assert.Nil(t, err)

	input = `
	-> Knot_A
	== Knot_A
	this is knot a -> stitch_a
	= stitch_a
	this is stitch a -> end
	== Knot_A
	-> end
	`

	story = Default()
	err = story.Parse(input)
	assert.Contains(t, err.Error(), "conflict knot name")
}

func TestKnotDivert(t *testing.T) {
	input := `
	this is a basic parsing test
	-> Knot_A
	== Knot_A
	this is knot a -> stitch_a
	= stitch_a
	this is stitch a -> Knot_B.stitch_b
	== Knot_B
	= stitch_a
	this is stitch a -> end
	= stitch_b
	-> stitch_a
	`

	story := Default()
	err := story.Parse(input)
	ctx := NewContext()

	_, err = story.Resume(ctx)
	assert.Nil(t, err)
}

func TestStitchParsing(t *testing.T) {
	input := `
	-> stitch_a
	= stitch_a
	-> end
	`

	story := Default()
	err := story.Parse(input)
	assert.Contains(t, err.Error(), "can not find the knot")

	input = `
	-> knot_a.stitch_a
	== knot_a
	hello
	= stitch_a
	-> end
	= stitch_a
	-> end
	`
	story = Default()
	err = story.Parse(input)
	assert.Contains(t, err.Error(), "conflict stitch")
}
