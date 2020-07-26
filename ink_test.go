package goink

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoryParse(t *testing.T) {
	input := `
	Once upon a time,
	There is a rabbit.

	* [Chase the rabbit]
	  ** [ABC]
	  ** [DEF]
	* [Shoot the rabbit]
	  ** [GHI]
	  ** [JKL]
	  ** [MNO]
	* [Do nothing] -> Knot_1 # This is a Tag // This is a comment

	=== Knot_1 ===
	  This is the knot_1 content.
	`
	contents := strings.Split(input, "\n")

	s := NewStory()
	s.start = &Inline{s: s}

	for _, line := range contents {
		Parse(s, line)
	}

	// plain text
	s.Reset()
	s.Next()
	assert.Equal(t, "Once upon a time,", s.current.(*Inline).raw)

	s.Next()
	s.Next()

	// choices
	_, err := s.Next()
	assert.Equal(t, "cannot go next", err.Error())

	s.Select(1)
	assert.Equal(t, "[Chase the rabbit]", s.current.(*Inline).text)

	s.Next()
	s.Select(2)
	assert.Equal(t, "[DEF]", s.current.(*Inline).text)

	end, err := s.Next()
	assert.Nil(t, end)

	s.Reset()
	s.Next()
	s.Next()
	s.Next()
	s.Select(3)
	assert.Equal(t, "[Do nothing] ", s.current.(*Inline).text)

	n, err := s.Next()
	t.Log(n)
	assert.Equal(t, "cannot go next", err.Error())

	// knot
	assert.NotNil(t, s.FindKnot("Knot_1"))
	assert.Equal(t, "Knot_1", s.knots[0].name)
}

func TestInlineParse(t *testing.T) {
	i := NewInline("This an inline. // This is a comment. ")
	assert.Equal(t, "This an inline. ", i.text)
	assert.Equal(t, "This is a comment.", i.comment)

	i = NewInline("* [Do nothing] -> Knot_1 #This is another tag # This is a Tag // This is a comment")
	assert.Equal(t, "Knot_1", i.divert)
	assert.Equal(t, "This is another tag", i.tags[1]) // index of tags is reversed
	assert.Equal(t, 2, len(i.tags))
}
