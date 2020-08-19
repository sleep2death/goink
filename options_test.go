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
	assert.Equal(t, "Opt A", sec.opts[0])
	assert.False(t, sec.end)

	_, err = story.Pick(ctx, 0) // pick Opt A
	assert.Nil(t, err)

	ctx = NewContext() // start with new context
	sec, err = story.Resume(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "Opt B", sec.opts[1])
	assert.False(t, sec.end)

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
	assert.Equal(t, "Opt ABC", sec.opts[0])
	assert.False(t, sec.end)

	sec, err = story.Pick(ctx, 0)       // pick Opt A, and fall to 'gather'
	assert.Contains(t, sec.text, "DEF") //supressing text
	assert.NotContains(t, sec.text, "ABC")

	assert.Equal(t, true, sec.end)
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
	assert.Equal(t, 2, len(sec.opts))
}
