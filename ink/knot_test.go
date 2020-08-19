package ink

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

	ctx := NewContext()
	_, err = story.Resume(ctx)
	assert.Nil(t, err)
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
	assert.Nil(t, err)

	ctx := NewContext()
	_, err = story.Resume(ctx)
	assert.Nil(t, err)
}
