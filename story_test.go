package goink

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

func TestStoryLoad(t *testing.T) {
	input := `
	This is a basic parsing test. # TAG_A
	Story will read these lines one by one, # tag b
	And connect them togather... # tag c // comment
	-> END
	`
	story := Default()
	err := story.Parse(input)
	story.SetID("ABC")
	assert.Nil(t, err)

	ctx := NewContext()
	ctx.current = "invalid path"
	_, err = story.Resume(ctx)
	assert.Contains(t, err.Error(), "is not existed")

	_, err = story.Pick(ctx, 0)
	assert.Contains(t, err.Error(), "is not existed")

	ctx.current = "start"
	_, err = story.Pick(ctx, 0)
	assert.Contains(t, err.Error(), "is not Choices")

	ctx = NewContext()
	ctx.Vars()["start__i"] = "invalid vars"
	_, err = story.Resume(ctx)
	assert.Contains(t, err.Error(), "is not type of int")
}

func BenchmarkBasicStoryParsing(b *testing.B) {
	input := `
	This is a basic parsing test. # TAG_A
	Story will read these lines one by one, # tag b
	And connect them togather... # tag c // comment
	-> END
	`
	for i := 0; i < b.N; i++ {
		story := Default()
		if err := story.Parse(input); err != nil {
			panic(err)
		}
	}
}

func BenchmarkComplexStoryParsing(b *testing.B) {
	input := `
    Hello
	-> Knot
	== Knot
	this is a knot content.
	* {knot > 0} Opt A
	  opt a content -> Knot
	* Opt B -> knot
	* Opt C
	- (gather) gather -> END
	== Knot_B
	this is a knot content.
	* {knot > 0} Opt A
	  opt a content -> Knot
	* Opt B -> knot
	* Opt C
	- (gather) gather -> END
	== Knot_C
	this is a knot content.
	* {knot > 0} Opt A
	  opt a content -> Knot
	* Opt B -> knot
	* Opt C
	- (gather) gather -> END
	`
	for i := 0; i < b.N; i++ {
		story := Default()
		if err := story.Parse(input); err != nil {
			b.Log(err)
		}
	}
}
