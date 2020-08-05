package goink

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Gather node of the choices
type Gather struct {
	Inline
	nesting int
}

var gatherReg = regexp.MustCompile(`^((-\s*)+)([^>].+)`)

// NewGather create and insert a new gather into story
func NewGather(s *Story, input string) error {
	res := gatherReg.FindStringSubmatch(input)
	if res != nil {
		nesting := len(strings.Join(strings.Fields(res[1]), ""))
		g := &Gather{nesting: nesting}
		g.raw = res[3]
		g.story = s

		obj := s.current
		var choices *Choices
		for obj != nil {
			if c, ok := obj.(*Choices); ok {
				if t := nesting - c.nesting; t == 0 {
					choices = c
					break
				}
			}

			obj = obj.Parent()
		}

		if choices != nil && choices.gather == nil {
			g.parent = choices.parent // set gather's grandpa to parent
			choices.gather = g
			s.current = g
			return nil
		}

		return errors.Errorf("cannot find the choice for the g nesting of the ather %s", input)
	}

	return ErrNotMatch
}

// Next content of the gather
func (g *Gather) Next() InkObj {
	if g.next != nil {
		return g.next
	}

	obj := g.parent

	for obj != nil {
		if c, ok := obj.(*Choices); ok {
			if c.gather != nil && c.gather != g {
				return c.gather
			}
		}

		obj = obj.Parent()
	}

	return nil
}
