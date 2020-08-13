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

			g.path = choices.Path() + split + "g"
			s.objMap[g.path] = g

			if err := g.parseLabel(); err != nil {
				return err
			}

			return nil
		}

		return errors.Errorf("cannot find the choices of the gather: %s", input)
	}

	return ErrNotMatch
}

func (g *Gather) parseLabel() error {
	if res := lableReg.FindStringSubmatch(g.text); res != nil {
		label := strings.TrimSpace(res[1])
		if len(label) > 0 {
			if knot, stitch := g.story.FindContainer(g); stitch != nil {
				label = stitch.Path() + split + label
			} else if knot != nil {
				label = knot.Path() + split + label
			}

			if _, ok := g.story.objMap[label]; ok {
				return errors.Errorf("duplicated label: %s", label)
			}
			g.story.objMap[label] = g
			g.path = label
		}
		g.text = res[2]
	}

	return nil
}
