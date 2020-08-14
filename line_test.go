package goink

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInlineParse(t *testing.T) {
	s := NewStory()

	err := readLine(s, "--> Divert")
	assert.NotNil(t, err)

	err = readLine(s, "-> Divert")
	assert.Nil(t, err)
	assert.Equal(t, "Divert", s.c.(*line).divert)

	err = readLine(s, "This is a content. -> Divert #Tag A # TagB // Comment")
	assert.Nil(t, err)
	assert.Equal(t, "Divert", s.c.(*line).divert)
	assert.Equal(t, "TagB", s.c.(*line).tags[1])
	assert.Equal(t, "Tag A", s.c.(*line).tags[0])
	assert.Equal(t, s, s.c.(*line).Story())
	assert.True(t, len(s.c.(*line).comment) > 0)
}

func TestDivert(t *testing.T) {
	input := `
	Hello, world!
	-> Knot_A

	== Knot_A
	* Option A
		** Option A.1
		** Option A.2
		-- Gather A
		   Gather A content -> Stitch_A_a
	= Stitch_A_a
		* Option B
		* Option C
		- Final Gather A -> Knot_B.Stitch_B_b
	== Knot_B
	* Option B
		** Option B.1 -> Stitch_B_b
		** Option B.2
		-- Gather B  -> Stitch_B_b
	= Stitch_B_b
		* Option B
		* Option C
		- Final Gather B
	`
	s, err := Parse(input)

	if err != nil {
		t.Error(err)
		return
	}

	rand.Seed(time.Now().UnixNano())

	for s.next() != nil {
		switch s.c.(type) {
		case *line:
			t.Log(s.c.(*line).Render())
		case *opt:
			t.Log(s.c.(*opt).render(true))
		case *gather:
			t.Log(s.c.(*gather).Render())
		case *options:
			for _, o := range s.c.(*options).list() {
				t.Log("*", o.render(true))
			}

			// random select
			idx := rand.Intn(len(s.c.(*options).list()))
			s.choose(idx)
			t.Logf("Select [%d]", idx)
		}
	}

	assert.Equal(t, "Final Gather B", s.c.(*gather).raw)
}

func TestGlueParse(t *testing.T) {
	input := `
	Glue Test 1
	<>Glue Test 2
	Glue Test 3 <>
	<>Glue Test 4<>
	`

	s, err := Parse(input)

	if err != nil {
		t.Error(err)
		return
	}

	s.next()
	s.next()

	assert.True(t, s.current().(*line).glueStart)
	assert.False(t, s.current().(*line).glueEnd)

	s.next()
	assert.True(t, s.current().(*line).glueEnd)
	assert.False(t, s.current().(*line).glueStart)

	s.next()
	assert.True(t, s.current().(*line).glueEnd)
	assert.True(t, s.current().(*line).glueStart)
}

func TestDivertNavigation(t *testing.T) {
	input := `
	-> Knot_A.stitch_b.lable_g
	== Knot_A
	This is Knot A.
	= stitch_b
	This is stitch b. -> stitch_c.lable_g
	* opt a
	+ opt b
	- (lable_g) gather -> stitch_b
	= stitch_c
	+ (lable_o) opt a -> unkown-divert
	+ opt b
	- (lable_g) gather c-> lable_o
	`

	s, err := Parse(input)

	if err != nil {
		t.Error(err)
		return
	}

	s.next()
	s.next()
	assert.Equal(t, " gather ", s.current().(*gather).Render())

	s.next()
	assert.Equal(t, "stitch_b", s.current().(*Stitch).name)

	s.next()
	s.next()
	assert.Equal(t, " gather c", s.current().(*gather).Render())

	s.next()
	assert.Equal(t, " opt a ", s.current().(*opt).render(false))

	pf := assert.PanicTestFunc(func() {
		s.next()
	})

	assert.Panics(t, pf)
}
