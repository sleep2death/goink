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

func TestGatherParsing(t *testing.T) {
	input := `
	* Opt A
	  opt a content
	  ** Opt A.1
	  opt a.1 content
	* Opt B
	  opt b content
	* Opt C
	  opt c content
	- gather -> END
	`

	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	ctx := NewContext()
	sec, err := story.Resume(ctx)
	assert.Nil(t, err)
	assert.Empty(t, sec.text)

	_, err = story.Pick(ctx, 0) // select Opt-A
	assert.Nil(t, err)

	sec, err = story.Pick(ctx, 0) // select Opt-A.1
	assert.Nil(t, err)

	assert.Contains(t, sec.text, "opt a.1 content")
	assert.Contains(t, sec.text, "gather")

	input = `
	- gather -> END
	`

	story = Default()
	err = story.Parse(input)
	assert.NotNil(t, err)
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
}

func TestOnceOnlyOption(t *testing.T) {
	input := `
    Hello
	-> Knot
	== Knot
	this is a knot content.
	* Opt A
	  opt a content -> Knot
	* Opt B -> knot
	* Opt C
	- (gather) gather -> END
	`
	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	ctx := NewContext()
	sec, err := story.Resume(ctx)
	assert.Nil(t, err)
	assert.Contains(t, sec.text, "this is a knot content")
	assert.Equal(t, 3, len(sec.opts))

	sec, err = story.Pick(ctx, 0)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(sec.opts)) // opt a should removed from list

	sec, err = story.Pick(ctx, 0)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(sec.opts)) // opt b should removed from list
	assert.Contains(t, sec.text, "Opt B")
}

func TestConditionalOption(t *testing.T) {
	input := `
    Hello
	-> Knot
	== Knot
	this is a knot content.
	* {knot > 0}Opt A
	  opt a content -> Knot
	* {knot == 0}Opt B -> knot
	* Opt C
	- (gather) gather -> END
	`

	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	ctx := NewContext()
	sec, err := story.Resume(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(sec.opts))
}
