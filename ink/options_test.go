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
	`
	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	ctx := NewContext()

	sec, err := story.Resume(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "Opt A", sec.opts[0])
	assert.False(t, sec.end)

	sec, err = story.Pick(ctx, 0)
	assert.Nil(t, err)

	t.Log(sec.text)
}
