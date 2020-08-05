package goink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidOptionNesting(t *testing.T) {
	input := `
	Hello World,
	* This is Option_A
		** Nesting Option_A.1
		**** Invalid Nesting Option_A.1.1
	* This is Option_B
	`
	_, err := parse(input)
	assert.NotNil(t, err)
}

func TestChoicesFunctions(t *testing.T) {
	input := `
	Hello, World

	* This is Option_A
		** Nesting Option_A.1
		** Nesting Option_A.2
		-- Gather A

	* This is Option_B

	- Final Gather
	`
	s, err := parse(input)
	assert.Nil(t, err)

	s.Next()
	assert.Equal(t, "Hello, World", s.Current().(*Inline).Render())

	s.Next()
	assert.Panics(t, assert.PanicTestFunc(func() { s.Current().(*Choices).SetNext(nil) }))

	if choices, ok := s.Current().(*Choices); ok {
		// Choices can not go next, always return nil
		assert.Nil(t, s.Next())

		assert.Equal(t, 1, choices.Nesting())
		assert.Equal(t, 2, len(choices.Options()))
		assert.Nil(t, nil, choices.Select(3))

		assert.Equal(t, s, choices.Story())
	} else {
		t.Error("should be choices type")
	}

}
