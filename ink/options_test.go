package ink

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
