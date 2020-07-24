package goink

import (
	"regexp"

	"github.com/pkg/errors"
)

var knotReg = regexp.MustCompile(`^={2,}\s(\w+)`)

// Knot of the story
type Knot struct {
	name    string
	story   *Story
	content []Block
}

// Parent of the knot should always be story
func (k *Knot) Parent() Block {
	return k.story
}

// Content of the knot
func (k *Knot) Content() []Block {
	return k.content
}

func (k *Knot) parse(raw string) (blk Block, err error) {
	// -- container header --
	// == KNOT ==
	if knot := newKnot(raw); knot != nil {
		// handle the raw to story
		return k.story.parse(raw)
	}
	// +* CHOICE *+
	if choice := newChoice(raw); choice != nil {
		// diff from original ink, force check nesting level of choices
		// for better reading and writing
		if choice.nesting > 1 {
			return nil, errors.Errorf("original nesting of the knot choice should always be 1: %s", raw)
		}

		// choice.nesting = 1 // always be 2

		choice.parent = k
		k.content = append(k.content, choice)
		return choice, nil
	}

	// INLINE
	inline := newInline(raw)
	inline.parent = k
	k.content = append(k.content, inline)
	return k, nil
}

// newKnot generator
func newKnot(str string) *Knot {
	res := knotReg.FindStringSubmatch(str)
	if len(res) > 0 {
		return &Knot{name: res[1]}
	}
	return nil
}
