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
	assert.Equal(t, "\n[start]\n[end]", sec.text)
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

	ctx := &Context{current: "start"}
	ctx.vars = make(map[string]interface{})

	sec, err := story.Resume(ctx)
	assert.Nil(t, err)
	assert.Equal(t, true, sec.end)
	assert.Equal(t, "end", ctx.Current())
}
