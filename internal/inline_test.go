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
	  ** [GHI]
	  ** [JKL]
	  ** [MNO]
	* [Do nothing]
	`
	contents := strings.Split(input, "\n")

	s := NewStory()
	s.start = &PlainText{s: s}

	for _, line := range contents {
		err := Parse(s, line)
		assert.Error(t, err)
	}

	s.Reset()
	n, _ := s.Next()
	assert.Equal(t, "Once upon a time,", n.(*PlainText).raw)

	_, _ = s.Next()
	_, _ = s.Next()

	_, err := s.Next()
	assert.Equal(t, "cannot go next", err.Error())

	_, _ = s.Select(1)
	assert.Equal(t, "[Chase the rabbit]", s.current.(*PlainText).raw)

	_, _ = s.Next()
	_, _ = s.Select(2)
	assert.Equal(t, "[DEF]", s.current.(*PlainText).raw)

	end, err := s.Next()
	assert.Nil(t, err)
	assert.Nil(t, end)

	s.Reset()
	_, _ = s.Next()
	_, _ = s.Next()
	_, _ = s.Next()
	_, _ = s.Select(3)
	assert.Equal(t, "[Do nothing]", s.current.(*PlainText).raw)
}
