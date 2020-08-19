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

	env := make(map[string]interface{})
	env["Knot_A"] = 1
	env["intA"] = 2
	env["intB"] = -1

	b, err := condA.Bool(env)

	assert.Nil(t, err)
	assert.False(t, b)

	b, _ = condB.Bool(env)
	assert.False(t, b)

	b, err = condC.Bool(env)
	assert.NotNil(t, err)
	assert.False(t, b)

	b, err = condD.Bool(env)
	assert.Nil(t, err)
	assert.False(t, b)

	b, err = condE.Bool(env)
	assert.Nil(t, err)
	assert.True(t, b)

	b, err = condF.Bool(env)
	assert.NotNil(t, err)
	assert.False(t, b)
}
