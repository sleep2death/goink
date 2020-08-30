package goink

import (
	"strings"
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
	assert.Contains(t, sec.Text, "No.2")

	l1, _ := story.paths["start__i"].(*line)
	assert.Equal(t, 3, len(l1.tags))
	assert.Equal(t, "tag a", l1.tags[0])

	_, ok := story.paths["start__i__i"].(*line)
	assert.True(t, ok)
	// assert.True(t, l2.glueStart)
	// assert.True(t, l2.glueEnd)
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
	assert.Equal(t, 3, len(sec.Opts))

	sec, err = story.Pick(ctx, 1)
	assert.Nil(t, err)
	assert.NotContains(t, sec.Text, "knot b")
	assert.Contains(t, sec.Text, "stitch content")

	_, err = story.Pick(ctx, 0)
	assert.Nil(t, err)

	opt, ok := story.paths["knot_b__stitch_a__i__c__0"]
	assert.True(t, ok)
	assert.Equal(t, "knot_b__stitch_a__label", opt.Path())

	input = `
	* (duplicated_label) Opt A
	  opt a content
	* (duplicated_label)Opt B -> knot_b.stitch_a
	* Opt C
	- (gather) gather -> END
	`

	story = Default()
	err = story.Parse(input)
	assert.NotNil(t, err)
}

func TestDivertParsing(t *testing.T) {
	input := `go to invalid divert -> divert`

	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	ctx := NewContext()
	_, err = story.Resume(ctx)
	assert.Equal(t, "can not find the divert: divert ln: 1", err.Error())

	input = `go to invalid divert -> invalid divert`

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

func TestDivertJumping(t *testing.T) {
	input := `
	no next node available
	`

	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	ctx := NewContext()
	_, err = story.Resume(ctx)
	assert.NotNil(t, err)
}

func TestDivertValidation(t *testing.T) {
	input := `
	invalid divert name -> end []
	`

	story := Default()
	err := story.Parse(input)
	assert.NotNil(t, err)
}

func TestCommentParsing(t *testing.T) {
	input := `
	a line with comment // http://www.sina.com
	`

	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	t.Log(story.paths["start__i"].(*line).comment)
}

func TestGlueRendering(t *testing.T) {
	input := `
	this is a tail glue <>
	this is the second line.
	-> end
	`

	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	e := story.PostParsing()
	assert.Nil(t, e)

	ctx := NewContext()
	sec, errs := story.Resume(ctx)
	assert.Nil(t, errs)

	assert.Equal(t, 1, len(strings.Split(sec.Text, "\n")))
	t.Log(sec.Text)

	input = `
	this is a tail glue
	<> this is the second line.
	-> end
	`

	story = Default()
	err = story.Parse(input)
	assert.Nil(t, err)

	e = story.PostParsing()
	assert.Nil(t, e)

	ctx = NewContext()
	sec, errs = story.Resume(ctx)
	assert.Nil(t, errs)

	assert.Equal(t, 1, len(strings.Split(sec.Text, "\n")))
	t.Log(sec.Text)
}

func TestVariableParsing(t *testing.T) {
	input := `
	VAR a = 123
	VAR b = false
	var  c="hello world"
	`

	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	assert.Equal(t, story.vars["a"], 123)
	assert.Equal(t, story.vars["b"], false)
	assert.Equal(t, story.vars["c"], "hello world")
}
