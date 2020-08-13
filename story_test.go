package goink

import (
	"math/rand"
	"strings"
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
	s, err := parse(input)

	if err != nil {
		t.Error(err)
		return
	}

	rand.Seed(time.Now().UnixNano())

	for s.Next() != nil {
		// t.Log(s.current.Path())
		switch s.current.(type) {
		case *Inline:
			t.Log(s.current.(*Inline).render())
		case *Option:
			t.Log(s.current.(*Option).Render(false))
		case *Gather:
			t.Log(s.current.(*Gather).render())
		case *Choices:
			for _, o := range s.current.(*Choices).options() {
				t.Log("*", o.Render(true))
			}

			// random select
			idx := rand.Intn(len(s.current.(*Choices).options()))
			s.Select(idx)
			t.Logf("Select [%d]: %s", idx, s.current.(*Option).Render(false))
		}
	}

	assert.Equal(t, "Final Content", s.current.(*Inline).render())
	assert.Equal(t, "Knot_A__stitch_a__i", s.current.Path())
	assert.Equal(t, s.objMap["Knot_A__stitch_a__i"], s.current)
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
	s, err := parse(input)

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
	s, err := parse(input)

	if err != nil {
		t.Error(err)
		return
	}

	s.Next()
	s.Next()
	s.Next()

	assert.Equal(t, "Knot_A", s.Current().Path())

	s.Next()
	s.Next()
	s.Next()

	o := s.Select(8) // invalid test
	assert.Nil(t, o)

	o = s.Select(0)
	assert.Equal(t, "Knot_A__Stitch_A__c__0", o.Path())

	state := s.Save()
	assert.Equal(t, "Knot_A__Stitch_A__c__0", state.Path())
	assert.Equal(t, 1, state.Count()["Knot_A__Stitch_A__c__0"])

	// create a new story from the same source
	ss, err := parse(input)
	assert.Nil(t, err)
	err = ss.Load(state)

	assert.Nil(t, err)
	assert.Equal(t, "Knot_A__Stitch_A__c__0", s.Current().Path())
	assert.Equal(t, 1, s.objCount[s.Current().Path()])

	ss.Next()
	ss.Next()
	ss.Next()
	assert.Equal(t, "Knot_B__Stitch_A__i", ss.Current().Path())

	state.path = "invalid path"
	err = ss.Load(state)
	assert.NotNil(t, err)
}

func parse(input string) (*Story, error) {
	contents := strings.Split(input, "\n")

	s := NewStory()

	for _, line := range contents {
		if err := s.parseLine(line); err != nil {
			return nil, err
		}
	}

	s.Reset()
	return s, nil
}
