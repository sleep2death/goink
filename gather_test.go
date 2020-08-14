package goink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGather(t *testing.T) {
	input := `
	Hello World,
	* Option A
	* Option B
	- --> Knot_A
	# Knot_A
	A content here.
	`
	_, err := Parse(input)
	assert.NotNil(t, err)
}

func TestGatherLableParsing(t *testing.T) {
	input := `
	Hello World,
	* (lable_a) Option A
	* Option B
	- (lable_a)-> Knot_A
	# Knot_A
	A content here.
	`
	_, err := Parse(input)
	assert.NotNil(t, err)

	input = `
	Hello World,
	-> Knot_A
	== Knot_A
	  -> Stitch_A
	  = Stitch_A
	  A content here.
	  * Option A
	  * Option B
	  - (lable_g)-> Knot_A
	`

	s, err := Parse(input)
	assert.Nil(t, err)

	s.next()
	s.next()
	s.next()
	s.next()
	s.next()
	s.next()
	s.next()
	s.choose(0)
	t.Log(s.current())
	s.next()

	assert.Equal(t, "Knot_A__Stitch_A__lable_g", s.current().(*gather).Path())
}
