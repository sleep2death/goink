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
		** Option B.1
		** Option B.2
		-- Gather B
		   Gather B content -> Stitch_B_b
	= Stitch_B_b
		* Option B
		* Option C
		- Final Gather B
	`
	s, err := parse(input)

	if err != nil {
		t.Error(err)
		return
	}

	rand.Seed(time.Now().UnixNano())

	for s.Next() != nil {
		switch s.current.(type) {
		case *Inline:
			t.Log(s.current.(*Inline).Render())
		case *Option:
			t.Log(s.current.(*Option).Render())
		case *Gather:
			t.Log(s.current.(*Gather).Render())
		case *Choices:
			for _, o := range s.current.(*Choices).options {
				t.Log("*", o.raw)
			}

			// random select
			idx := rand.Intn(len(s.current.(*Choices).options))
			s.current.(*Choices).Select(idx)
			t.Logf("Select [%d]", idx)
		}
	}

	assert.Equal(t, "Final Gather B", s.current.(*Gather).raw)
}
