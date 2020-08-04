package goink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInlineParse(t *testing.T) {
	s := NewStory()

	err := NewInline(s, "--> Divert")
	assert.NotNil(t, err)

	err = NewInline(s, "-> Divert")
	assert.Nil(t, err)
	assert.Equal(t, "Divert", s.current.(*Inline).divert)

	err = NewInline(s, "This is a content. -> Divert #Tag A # TagB // Comment")
	assert.Nil(t, err)
	assert.Equal(t, "Divert", s.current.(*Inline).divert)
	assert.Equal(t, "TagB", s.current.(*Inline).tags[1])
	assert.Equal(t, "Tag A", s.current.(*Inline).tags[0])
	assert.True(t, len(s.current.(*Inline).comment) > 0)
}
