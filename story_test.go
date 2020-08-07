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
				*** Option A.3.1
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
        - Final Gather 2
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
				t.Log("*", o.Render())
			}

			// random select
			idx := rand.Intn(len(s.current.(*Choices).options))
			s.current.(*Choices).Select(idx)
			t.Logf("Select [%d]", idx)
		}
	}

	assert.Equal(t, "Final Gather 2", s.current.(*Gather).raw)
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
	assert.Equal(t, "root.k_Knot_A", s.knots[0].Path())

	// s for "stitch"
	assert.Equal(t, "root.k_Knot_A.s_Stitch_A", s.knots[0].stitches[0].Path())
	assert.Equal(t, "root.k_Knot_B.s_Stitch_A", s.knots[1].stitches[0].Path())

	assert.Equal(t, s.objMap["root.k_Knot_B.s_Stitch_A"], s.knots[1].stitches[0])
}

func parse(input string) (*Story, error) {
	contents := strings.Split(input, "\n")

	s := NewStory()

	for _, line := range contents {
		if err := s.Parse(line); err != nil {
			return nil, err
		}
	}

	s.Reset()

	return s, nil
}
