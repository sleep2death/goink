package internal

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
	  ** [ABC]
	  ** [DEF]
	  ** [GHI]
	* [Do nothing]
	`
	contents := strings.Split(input, "\n")

	s := NewStory()
	s.start = &PlainText{s: s}

	for _, line := range contents {
		Parse(s, line)
	}

	s.Reset()
	s.Next()
	assert.Equal(t, "Once upon a time,", s.current.(*PlainText).raw)

	s.Next()
	s.Next()

	_, err := s.Next()
	assert.Equal(t, "cannot go next", err.Error())

	s.Select(1)
	assert.Equal(t, "[Chase the rabbit]", s.current.(*PlainText).raw)

	s.Next()
	s.Select(2)
	assert.Equal(t, "[DEF]", s.current.(*PlainText).raw)

	end, err := s.Next()
	assert.Nil(t, end)
}
