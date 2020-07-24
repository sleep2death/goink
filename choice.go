package goink

import (
	"regexp"

	"github.com/pkg/errors"
)

// ChoiceType - plus or star
type ChoiceType int

const (
	// CPlus - Choice Type Plus (+)
	CPlus ChoiceType = iota
	// CStar - Choice Type Star (*)
	CStar
)

var choiceReg = regexp.MustCompile(`((^\++)|(^\*+))\s(.+)`)

// Choice content
type Choice struct {
	parent  Block
	nesting int
	content []Block

	ct  ChoiceType
	raw string
}

// Parent of the choice
func (c *Choice) Parent() Block {
	return c.parent
}

// Content of the choice
func (c *Choice) Content() []Block {
	return c.content
}

func (c *Choice) parse(raw string) (blk Block, err error) {
	// -- container header --
	// == KNOT ==
	if knot := newKnot(raw); knot != nil {
		// handle the raw to story
		return c.parent.parse(raw)
	}
	// +* CHOICE *+
	if choice := newChoice(raw); choice != nil {
		// diff from original ink, force check nesting level of choices
		// for better reading and writing
		if (choice.nesting - c.nesting) > 1 {
			return nil, errors.Errorf("child nesting error: %s", raw)
		}

		if c.nesting < choice.nesting {
			// insert inline content of the choice
			inline := newInline(c.raw)
			inline.parent = c
			c.content = append(c.content, inline)

			choice.parent = c
			c.content = append(c.content, choice)
			return choice, nil
		}

		return c.parent.parse(raw)
	}

	// INLINE
	inline := newInline(raw)
	inline.parent = c
	c.content = append(c.content, inline)

	return c, nil
}

// newChoice generator
func newChoice(str string) *Choice {
	res := choiceReg.FindStringSubmatch(str)
	if res != nil {
		if len(res[2]) > 0 {
			return &Choice{nesting: len(res[2]), ct: CPlus, raw: res[4]}
		}

		if len(res[3]) > 0 {
			return &Choice{nesting: len(res[3]), ct: CStar, raw: res[4]}
		}
	}
	return nil
}
