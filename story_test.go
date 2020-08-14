package goink

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChoicesNesting(t *testing.T) {
	input := `
	Hello, world!
		* Option A
			** Option A.1
			** Option A.2
			-- Gather A
               Gather A cotent
			** Option A.3
				A.3 Content
				*** Option A.3[.]1
					A.3.1 Content
			** Option A.4
		* Option B
			** Option B.1
			   B.1 Content
			** Option B.2
			-- Gather B
        - Final Gather 1
		* Option C
		  Option C Content
        - Final Gather 2 -> Knot_A.stitch_a
		== Knot_A
		= stitch_a
		   Final Content
	`
	s, err := Parse(input)

	if err != nil {
		t.Error(err)
		return
	}

	rand.Seed(time.Now().UnixNano())

	for s.next() != nil {
		// t.Log(s.current.Path())
		switch s.c.(type) {
		case *line:
			t.Log(s.c.(*line).Render())
		case *opt:
			t.Log(s.c.(*opt).render(false))
		case *gather:
			t.Log(s.c.(*gather).Render())
		case *options:
			for _, o := range s.c.(*options).list() {
				t.Log("*", o.render(true))
			}

			// random select
			idx := rand.Intn(len(s.c.(*options).list()))
			s.choose(idx)
			t.Logf("Select [%d]: %s", idx, s.c.(*opt).render(false))
		}
	}

	assert.Equal(t, "Final Content", s.c.(*line).Render())
	assert.Equal(t, "Knot_A__stitch_a__i", s.c.Path())
	assert.Equal(t, s.paths["Knot_A__stitch_a__i"], s.c)
}

func TestInkObjPath(t *testing.T) {
	input := `
	Hello, world!
	-> Knot_A

	== Knot_A ==
	This is knot_a content.
		= Stitch_A
		This is stitch_a content.
	== Knot_B ==
	This is knot_a content.
		= Stitch_A
		This is stitch_a content.
	`
	s, err := Parse(input)

	if err != nil {
		t.Error(err)
		return
	}

	// k for "knot"
	assert.Equal(t, "Knot_A", s.knots[0].Path())

	// s for "stitch"
	assert.Equal(t, "Knot_A__Stitch_A", s.knots[0].stitches[0].Path())
	assert.Equal(t, "Knot_B__Stitch_A", s.knots[1].stitches[0].Path())

	// assert.Equal(t, s.objMap["r.k_Knot_B.s_Stitch_A"], s.knots[1].stitches[0])
}

func TestStorySave(t *testing.T) {
	input := `
	Hello, world!
	-> Knot_A

	== Knot_A ==
	This is knot_a content. -> Stitch_A
		= Stitch_A
		* Option A
		* Option B
		- This is stitch_a content. -> Knot_B.Stitch_A
	== Knot_B ==
	This is knot_a content.
		= Stitch_A
		This is stitch_a content.
	`
	s, err := Parse(input)

	if err != nil {
		t.Error(err)
		return
	}

	s.next()
	s.next()
	s.next()

	assert.Equal(t, "Knot_A", s.current().Path())

	s.next()
	s.next()
	s.next()

	o := s.choose(8) // invalid test
	assert.Nil(t, o)

	o = s.choose(0)
	assert.Equal(t, "Knot_A__Stitch_A__c__0", o.Path())

	state := s.Save()
	assert.Equal(t, "Knot_A__Stitch_A__c__0", state.Path())
	assert.Equal(t, 1, state.Count()["Knot_A__Stitch_A__c__0"])

	// create a new story from the same source
	ss, err := Parse(input)
	assert.Nil(t, err)
	err = ss.Load(state)

	assert.Nil(t, err)
	assert.Equal(t, "Knot_A__Stitch_A__c__0", s.current().Path())
	assert.Equal(t, 1, s.vars[s.current().Path()])

	ss.next()
	ss.next()
	ss.next()
	assert.Equal(t, "Knot_B__Stitch_A__i", ss.current().Path())

	state.path = "invalid path"
	err = ss.Load(state)
	assert.NotNil(t, err)
}
