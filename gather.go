package goink

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Gather node of the choices
type Gather struct {
	*Inline
	nesting int
}

var gatherReg = regexp.MustCompile(`^((-\s*)+)([^>].+)`)

// NewGather create and insert a new gather into story
func NewGather(s *Story, input string) error {
	res := gatherReg.FindStringSubmatch(input)
	if res != nil {
		nesting := len(strings.Join(strings.Fields(res[1]), ""))
		// g := &Gather{nesting: nesting}
		i, err := CreateNewline(res[3])
		if err != nil {
			return err
		}

		g := &Gather{Inline: i, nesting: nesting}
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

		return errors.Errorf("cannot find the choice of the gather: %s", input)
	}

	return ErrNotMatch
}

// Next content of the gather
func (g *Gather) Next() InkObj {
	if g.divert != "" {
		return g.Inline.Next()
	}

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
