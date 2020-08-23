package goink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptionsParse(t *testing.T) {
	input := `
	* Opt A
	    This is Option A -> END
	* Opt B
		This is Option B
		* * Opt C
		This is Option C
		-> END
	`
	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	ctx := NewContext()

	sec, err := story.Resume(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "Opt A", sec.Opts[0])
	assert.False(t, sec.End)

	_, err = story.Pick(ctx, 0) // pick Opt A
	assert.Nil(t, err)

	ctx = NewContext() // start with new context
	sec, err = story.Resume(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "Opt B", sec.Opts[1])
	assert.False(t, sec.End)

	_, err = story.Pick(ctx, 1) // pick Opt B
	assert.Nil(t, err)

	_, err = story.Pick(ctx, 0)
	assert.Nil(t, err)
}

func TestGatherOfOptions(t *testing.T) {
	input := `
	* Opt [ABC]DEF
	    This is Option A
	* Opt B
		This is Option B
		* * Opt C
		This is Option C
	- gather -> End
	`
	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	ctx := NewContext()

	sec, err := story.Resume(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "Opt ABC", sec.Opts[0])
	assert.False(t, sec.End)

	sec, err = story.Pick(ctx, 0)       // pick Opt A, and fall to 'gather'
	assert.Contains(t, sec.Text, "DEF") //supressing text
	assert.NotContains(t, sec.Text, "ABC")

	assert.Equal(t, true, sec.End)
	assert.Nil(t, err)
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
	assert.Contains(t, sec.Text, "this is a knot content")
	assert.Equal(t, 3, len(sec.Opts))

	sec, err = story.Pick(ctx, 0)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(sec.Opts)) // opt a should removed from list

	sec, err = story.Pick(ctx, 0)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(sec.Opts)) // opt b should removed from list
	assert.Contains(t, sec.Text, "Opt B")
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
	assert.Empty(t, sec.Text)

	_, err = story.Pick(ctx, 0) // select Opt-A
	assert.Nil(t, err)

	sec, err = story.Pick(ctx, 0) // select Opt-A.1
	assert.Nil(t, err)

	assert.Contains(t, sec.Text, "opt a.1 content")
	assert.Contains(t, sec.Text, "gather")

	input = `
	- gather -> END
	`

	story = Default()
	err = story.Parse(input)
	assert.NotNil(t, err)

	input = `
	* a
	* b
	- (illegal label)gather -> END
	`

	story = Default()
	err = story.Parse(input)
	assert.NotNil(t, err)
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
	assert.Equal(t, 2, len(sec.Opts))
}

func TestStickyOption(t *testing.T) {
	input := `
    Hello
	-> Knot
	== Knot
	this is a knot content.
	+ {knot > 0}Opt A
	  opt a content -> Knot
	+ {knot == 0}Opt B -> knot
	+ Opt C
	- (gather) gather -> END
	`

	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	ctx := NewContext()
	sec, err := story.Resume(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(sec.Opts))

	sec, err = story.Resume(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(sec.Opts))
}

func TestOpitionParenting(t *testing.T) {
	input := `
	* Hello A
	** Hello B
	**** Hello C
	`

	story := Default()
	err := story.Parse(input)
	assert.Contains(t, err.Error(), "wrong nesting")

	input = `
	* { fff > } Hello A -> END
	** Hello B -> end
	`

	story = Default()
	err = story.Parse(input)
	assert.NotNil(t, err)
}

func TestNestingGatherAndLabel(t *testing.T) {
	input := `
	* (label_a)Hello A -> knot_a
	** (label_b)Hello B -> knot_b
	*** (label_c)Hello C -> knot_c

	== knot_a
	# knot a tag
	* (label_a) a
	* (label_b) b
	* (label_c) c
	== knot_b
	# knot b tag
	* (label_a) a -> stitch_a
	== stitch
	* (label_b) b
	* (label_c) c
	== knot_c
	# knot c tag
	* (label_a) a
	* (label_b) b
	* (label_c) c
	`

	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)
}

func TestOptionPostParsing(t *testing.T) {
	input := `* (label_a)Hello A -> knot_a
	** (label_b)Hello B -> knot_b
	*** (label_c)Hello C -> knot_cz`

	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	err = story.PostParsing()
	assert.Contains(t, err.Error(), "can not find the divert")
	assert.Contains(t, err.Error(), "ln: 1")
}
