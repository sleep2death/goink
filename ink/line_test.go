package ink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineParsing(t *testing.T) {
	input := `
	this is a basic line parsing test: # tag a # tag b#tag c
	<> sentence No.1 <>
	sentence No.2 -> End // comments...
	`

	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	ctx := NewContext()
	sec, err := story.Resume(ctx)
	assert.Nil(t, err)
	assert.Contains(t, sec.text, "No.2")

	l1, _ := story.paths["start__i"].(*line)
	assert.Equal(t, 3, len(l1.tags))
	assert.Equal(t, "tag a", l1.tags[0])

	l2, ok := story.paths["start__i__i"].(*line)
	assert.True(t, ok)
	assert.True(t, l2.glueStart)
	assert.True(t, l2.glueEnd)
}

func TestLabelParsing(t *testing.T) {
	input := `
	* (label) Opt A
	  opt a content
	* Opt B
	* Opt C
	- (gather) gather -> END
	`

	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	opts, ok := story.paths["start__c"].(*options)
	assert.True(t, ok)
	assert.Equal(t, "label", opts.opts[0].path)
	assert.Equal(t, "gather", opts.gather.path)

	input = `
	* (illegal label) Opt A
	  opt a content
	* Opt B
	* Opt C
	- (gather) gather -> END
	`
	story = Default()
	err = story.Parse(input)
	assert.NotNil(t, err)

	input = `
	-> knot_a
	== knot_a
		* (label) Opt A
		  opt a content
		* Opt B -> knot_b.stitch_a
		* Opt C
		- (gather) gather -> END
	== knot_b
		knot b content
		-> stitch_a
		= stitch_a
			stitch content here...
			* (label)Opt A
			  -> END
			* Opt B
			* Opt C
	`
	story = Default()
	err = story.Parse(input)
	assert.Nil(t, err)

	ctx := NewContext()
	sec, err := story.Resume(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(sec.opts))

	sec, err = story.Pick(ctx, 1)
	assert.Nil(t, err)
	assert.NotContains(t, sec.text, "knot b")
	assert.Contains(t, sec.text, "stitch content")

	_, err = story.Pick(ctx, 0)
	assert.Nil(t, err)

	opt, ok := story.paths["knot_b__stitch_a__i__c__0"]
	assert.True(t, ok)
	assert.Equal(t, "knot_b__stitch_a__label", opt.Path())
}

func TestDivertParsing(t *testing.T) {
	input := `
	go to invalid divert -> divert
	`

	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	ctx := NewContext()
	_, err = story.Resume(ctx)
	assert.Equal(t, "can not find the divert: <divert>", err.Error())

	input = `
	go to invalid divert -> invalid divert
	`

	story = Default()
	err = story.Parse(input)
	assert.NotNil(t, err)

	input = `
	go to invalid divert -> invalid.divert..
	`

	story = Default()
	err = story.Parse(input)
	assert.NotNil(t, err)
}
