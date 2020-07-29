package goink

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoryParse(t *testing.T) {
	input := `
	Once upon a time,
	There is a rabbit.

	* [Chase the rabbit]
	  *** [ABC] // weird nesting struggling
	  ** [DEF]
	* [Shoot the rabbit]
	  ** [GHI]
	  ** [JKL]
	  ** [MNO]
	* [Do nothing] -> Knot_1 # This is a Tag // This is a comment

	=== Knot_1 ===
	  This is the knot_1 content.
	  -> Stitch_1
	  = Stitch_1
	    Stitch content here.
	    go to the non-exist stitch. -> Stitch_2
	`
	contents := strings.Split(input, "\n")

	s := NewStory()
	s.start = &Inline{s: s}

	for _, line := range contents {
		if err := Parse(s, line); err != nil {
			t.Error(err)
		}
	}

	// plain text
	s.Reset()
	n, _ := s.Next()
	assert.Equal(t, "Once upon a time,", n.(*Inline).raw)

	_, _ = s.Next()
	_, _ = s.Next()

	// choices
	_, err := s.Next()
	assert.Equal(t, "cannot go next: 5", err.Error())

	n, _ = s.Select(1)
	assert.Equal(t, "[Chase the rabbit]", n.(*Inline).text)
	assert.Equal(t, s, n.Story())

	_, _ = s.Next()
	n, _ = s.Select(2)
	assert.Equal(t, "[DEF]", n.(*Inline).text)

	end, err := s.Next()
	assert.Equal(t, "cannot go next: 7", err.Error())
	assert.Nil(t, end)

	s.Reset()
	_, _ = s.Next()
	_, _ = s.Next()
	_, _ = s.Next()
	assert.Equal(t, s, s.current.Story())
	n, _ = s.Select(3)
	// t.Log(n)
	assert.Equal(t, "[Do nothing] ", n.(*Inline).text)

	// divert
	_, _ = s.Next()
	assert.Equal(t, "This is the knot_1 content.", s.current.(*Inline).text)

	// knot
	assert.NotNil(t, s.FindKnot("Knot_1"))
	assert.Equal(t, "Knot_1", s.knots[0].name)

	// divert
	_, _ = s.Next()
	assert.Equal(t, "", s.current.(*Inline).text)
	assert.Equal(t, s.FindKnot("Knot_1"), s.current.(*Inline).k)

	// stitch
	_, _ = s.Next()
	assert.Equal(t, "Stitch content here.", s.current.(*Inline).text)
	assert.Equal(t, s, s.current.Story())

	n, _ = s.Next()
	assert.Equal(t, "Stitch_2", n.(*Inline).divert)
	_, err = s.Next()
	assert.Equal(t, "cannot go next: 19", err.Error())
}

func TestInlineParse(t *testing.T) {
	i := NewInline("This an inline. // This is a comment. ")
	assert.Equal(t, "This an inline. ", i.text)
	assert.Equal(t, "This is a comment.", i.comment)

	i = NewInline("* [Do nothing] -> Knot_1 #This is another tag # This is a Tag // This is a comment")
	assert.Equal(t, "Knot_1", i.divert)
	assert.Equal(t, "This is another tag", i.tags[0]) // index of tags is reversed
	assert.Equal(t, 2, len(i.tags))

	i = NewInline("-> Knot_1")
	assert.Equal(t, "", i.text)
	assert.Equal(t, "Knot_1", i.divert)

	i = NewInline("#TAG_1 #TAG_2")
	assert.Equal(t, "", i.text)
	assert.Equal(t, "TAG_1", i.tags[0])
}

func TestGatherParse(t *testing.T) {
	input := `
	- 	"Well, Poirot? Murder or suicide?"
		*	"Murder!"
		 	"And who did it?"
			* * 	"Detective-Inspector Japp!"
			* * 	"Captain Hastings!"
			* * 	"Myself!"
			- -     "You must be joking!"
			* * 	"Mon ami, I am deadly serious."
			* *		"If only..."
		* 	"Suicide!"
			"Really, Poirot? Are you quite sure?"
			* * 	"Quite sure."
			* *		"It is perfectly obvious."
		-	Mrs. Christie lowered her manuscript a moment. The rest of the writing group sat, open-mouthed.
	`
	contents := strings.Split(input, "\n")

	s := NewStory()
	s.start = &Inline{s: s}

	for _, line := range contents {
		if err := Parse(s, line); err != nil {
			t.Error(err)
		}
	}

	// plain text
	s.Reset()
	n, _ := s.Next()
	assert.Equal(t, "\"Well, Poirot? Murder or suicide?\"", n.(*Gather).raw)
	_, _ = s.Next()
	n, _ = s.Select(1)
	assert.Equal(t, "\"Murder!\"", n.(*Inline).raw)
	n, _ = s.Next()
	assert.Equal(t, "\"And who did it?\"", n.(*Inline).raw)
	_, _ = s.Next()
	n, _ = s.Select(3)
	assert.Equal(t, "\"Myself!\"", n.(*Inline).raw)
}
