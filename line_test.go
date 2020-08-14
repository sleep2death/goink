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
	assert.Equal(t, "Divert", s.current.(*line).divert)

	err = readLine(s, "This is a content. -> Divert #Tag A # TagB // Comment")
	assert.Nil(t, err)
	assert.Equal(t, "Divert", s.current.(*line).divert)
	assert.Equal(t, "TagB", s.current.(*line).tags[1])
	assert.Equal(t, "Tag A", s.current.(*line).tags[0])
	assert.Equal(t, s, s.current.(*line).Story())
	assert.True(t, len(s.current.(*line).comment) > 0)
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

	for s.Next() != nil {
		switch s.current.(type) {
		case *line:
			t.Log(s.current.(*line).render())
		case *Option:
			t.Log(s.current.(*Option).Render(true))
		case *gather:
			t.Log(s.current.(*gather).render())
		case *Choices:
			for _, o := range s.current.(*Choices).options() {
				t.Log("*", o.Render(true))
			}

			// random select
			idx := rand.Intn(len(s.current.(*Choices).options()))
			s.Select(idx)
			t.Logf("Select [%d]", idx)
		}
	}

	assert.Equal(t, "Final Gather B", s.current.(*gather).raw)
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

	s.Next()
	s.Next()

	assert.True(t, s.Current().(*line).glueStart)
	assert.False(t, s.Current().(*line).glueEnd)

	s.Next()
	assert.True(t, s.Current().(*line).glueEnd)
	assert.False(t, s.Current().(*line).glueStart)

	s.Next()
	assert.True(t, s.Current().(*line).glueEnd)
	assert.True(t, s.Current().(*line).glueStart)
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

	s.Next()
	s.Next()
	assert.Equal(t, " gather ", s.Current().(*gather).render())

	s.Next()
	assert.Equal(t, "stitch_b", s.Current().(*Stitch).name)

	s.Next()
	s.Next()
	assert.Equal(t, " gather c", s.Current().(*gather).render())

	s.Next()
	assert.Equal(t, " opt a ", s.Current().(*Option).Render(false))

	pf := assert.PanicTestFunc(func() {
		s.Next()
	})

	assert.Panics(t, pf)
}
