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
		i, err := CreateNewInline(res[3])
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

			g.path = choices.Path() + ".g"
			s.objMap[g.path] = g
			return nil
		}

		return errors.Errorf("cannot find the choice of the gather: %s", input)
	}

	return ErrNotMatch
}
