package ink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultStory(t *testing.T) {
	story := Default()
	assert.Equal(t, story, story.current.Story())

	ctx := story.save()
	assert.Equal(t, "start", ctx.Current())

	sec, err := story.Resume(&ctx)
	assert.Nil(t, err)
	assert.Equal(t, "", sec.text)
	assert.Equal(t, 2, len(sec.tags))
}

func TestBasicParse(t *testing.T) {
	input := `
	This is a basic parsing test. # TAG_A
	Story will read these lines one by one, # tag b
	And connect them togather... # tag c // comment
	-> END
	`
	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	ctx := NewContext()
	sec, err := story.Resume(ctx)
	assert.Nil(t, err)

	assert.Equal(t, true, sec.end)
	assert.Equal(t, "end", ctx.Current())
	assert.Equal(t, 5, len(sec.tags)) // 3 + start_tag + end_tag
}
