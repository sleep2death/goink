package goink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewExprc(t *testing.T) {
	condA, err := newExprc("not (Knot_A > 0)")
	assert.Nil(t, err)

	condB, err := newExprc("Knot_B == 0")
	assert.Nil(t, err)

	condC, err := newExprc("abc +  def")
	assert.Nil(t, err)

	condD, err := newExprc("(Knot_A > 0) and (Knot_B == 0)")
	assert.Nil(t, err)

	condE, err := newExprc("intA + intB")
	assert.Nil(t, err)

	condF, err := newExprc("'intA' + 'intB'")
	assert.Nil(t, err)

	_, err = newExprc("(Knot_A > 0 and (Knot_B == 0)")
	assert.NotNil(t, err)

	env := make(map[string]int)
	env["Knot_A"] = 1
	env["intA"] = 2
	env["intB"] = -1

	b, err := condA.Bool(env)

	assert.Nil(t, err)
	assert.False(t, b)

	b, err = condB.Bool(env)

	assert.Nil(t, err)
	assert.True(t, b)

	b, err = condC.Bool(env)
	assert.Nil(t, err)
	assert.False(t, b)

	b, err = condD.Bool(env)
	assert.Nil(t, err)
	assert.True(t, b)

	b, err = condE.Bool(env)
	assert.Nil(t, err)
	assert.True(t, b)

	b, err = condF.Bool(env)
	assert.NotNil(t, err)
	assert.False(t, b)
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
	s, err := Parse(input)

	if err != nil {
		t.Error(err)
		return
	}

	s.Next()
	s.Next()
	s.Next()

	c, ok := s.Current().(*options)
	assert.True(t, ok)

	// only one option will be displayed
	assert.Equal(t, 1, len(c.list()))

	condition0 := c.list()[0].condition
	assert.Equal(t, "Knot_A > 0", condition0.program.Source.Content())

	b, err := condition0.Bool(s.objCount)
	assert.Nil(t, err)
	assert.True(t, b)
}

func TestLableVisitCount(t *testing.T) {
	input := `
	-> Knot_A

	== Knot_A ==
		* { Knot_A > 0 }Option A
		* { Knot_B > 0 }Option B
		- (gather)Gather A -> Knot_B
	=== Knot_B ===
		+ {Knot_A.gather > 0} Option A
		+ {Knot_A.gather == 0} Option B
		- -> END
	`
	s, err := Parse(input)

	if err != nil {
		t.Error(err)
		return
	}

	s.Next()
	s.Next()
	s.Next()

	c, ok := s.Current().(*options)
	assert.True(t, ok)
	assert.Equal(t, 1, len(c.list()))

	s.Select(0)

	s.Next()
	s.Next()
	s.Next()

	c, ok = s.Current().(*options)
	assert.True(t, ok)

	// t.Log(s.objCount["Knot_A-gather"])
	t.Log(c.list()[0].condition.program.Source.Content())
	// t.Log(c.options()[0].condition.Bool(s.objCount))
	assert.Equal(t, 1, len(c.list()))
	assert.Equal(t, " Option A", c.list()[0].render(false))

	input = `
	-> Knot_A

	== Knot_A ==
		* { Knot_A > 0) }Option A // invalid condition
		* { Knot_B > 0 }Option B
		- (gather)Gather A -> Knot_B
	=== Knot_B ===
		+ {Knot_A.gather > 0} Option A
		+ {Knot_A.gather == 0} Option B
		- -> END
	`
	_, err = Parse(input)
	assert.NotNil(t, err)
}
