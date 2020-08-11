package goink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCondition(t *testing.T) {
	condA, err := NewCondition("not (Knot_A > 0)")
	assert.Nil(t, err)

	condB, err := NewCondition("Knot_B == 0")
	assert.Nil(t, err)

	condC, err := NewCondition("abc +  def")
	assert.Nil(t, err)

	condD, err := NewCondition("(Knot_A > 0) and (Knot_B == 0)")
	assert.Nil(t, err)

	_, err = NewCondition("(Knot_A > 0 and (Knot_B == 0)")
	assert.NotNil(t, err)

	env := make(map[string]int)
	env["Knot_A"] = 1

	b, err := condA.Bool(env)

	assert.Nil(t, err)
	assert.False(t, b)

	b, err = condB.Bool(env)

	assert.Nil(t, err)
	assert.True(t, b)

	_, err = condC.Bool(env)
	assert.NotNil(t, err)
	// assert.True(t, b)

	b, err = condD.Bool(env)
	assert.Nil(t, err)
	assert.True(t, b)
}

func TestKnotVisitCount(t *testing.T) {
	input := `
	-> Knot_A

	== Knot_A ==
		* { Knot_A > 0 }Option A
		* { Knot_B > 0 }Option A
		- Gather A -> Knot_B
	=== Knot_B ===
		This is Knot_B
		-> END
	`
	s, err := parse(input)

	if err != nil {
		t.Error(err)
		return
	}

	s.Next()
	s.Next()
	s.Next()

	c, ok := s.Current().(*Choices)
	assert.True(t, ok)

	// only one option will be displayed
	assert.Equal(t, 1, len(c.Options()))

	condition0 := c.Options()[0].condition
	assert.Equal(t, "Knot_A > 0", condition0.program.Source.Content())

	b, err := condition0.Bool(s.objCount)
	assert.Nil(t, err)
	assert.True(t, b)
}
