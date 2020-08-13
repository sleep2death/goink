package goink

import (
	"math/rand"
	"testing"
	"time"

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
	assert.Equal(t, s, s.current.(*Inline).Story())
	assert.True(t, len(s.current.(*Inline).comment) > 0)
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
		case *Inline:
			t.Log(s.current.(*Inline).render())
		case *Option:
			t.Log(s.current.(*Option).Render(true))
		case *Gather:
			t.Log(s.current.(*Gather).render())
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

	assert.Equal(t, "Final Gather B", s.current.(*Gather).raw)
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

	assert.True(t, s.Current().(*Inline).glueStart)
	assert.False(t, s.Current().(*Inline).glueEnd)

	s.Next()
	assert.True(t, s.Current().(*Inline).glueEnd)
	assert.False(t, s.Current().(*Inline).glueStart)

	s.Next()
	assert.True(t, s.Current().(*Inline).glueEnd)
	assert.True(t, s.Current().(*Inline).glueStart)
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
	assert.Equal(t, " gather ", s.Current().(*Gather).render())

	s.Next()
	assert.Equal(t, "stitch_b", s.Current().(*Stitch).name)

	s.Next()
	s.Next()
	assert.Equal(t, " gather c", s.Current().(*Gather).render())

	s.Next()
	assert.Equal(t, " opt a ", s.Current().(*Option).Render(false))

	pf := assert.PanicTestFunc(func() {
		s.Next()
	})

	assert.Panics(t, pf)
}
