package goink

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var choicesReg = regexp.MustCompile(`(^(\+\s*)+|^(\*\s*)+)(.+)`)

func NewOption(s *Story, input string) error {
	res := choicesReg.FindStringSubmatch(input)

	if res != nil {
		nesting := len(strings.Join(strings.Fields(res[1]), ""))

		o := &Option{}
		o.story = s
		o.raw = res[4]

		obj := s.current

		var choices *Choices
		// var gather *Gather

		for obj != nil {
			if g, ok := obj.(*Gather); ok {
				if t := nesting - g.nesting; t == 0 {
					break
				}
			}

			if c, ok := obj.(*Choices); ok {
				if t := nesting - c.nesting; t >= 0 {
					if t > 1 {
						return errors.Errorf("wrong nesting of the option: %s", input)
					} else if t == 0 {
						choices = c
					}
					break
				}
			}

			obj = obj.Parent()
		}
		if choices == nil {
			choices = &Choices{story: s, parent: s.current, nesting: nesting}
			s.current.SetNext(choices)
		}
		choices.options = append(choices.options, o)

		// s.current.SetNext(o)
		o.parent = choices
		s.current = o
		return nil
	}

	return NotMatch
}

type Choices struct {
	story   *Story
	parent  InkObj
	options []*Option

	gather  *Gather
	nesting int
}

func (c *Choices) Story() *Story {
	return c.story
}

func (c *Choices) Parent() InkObj {
	return c.parent
}

func (c *Choices) SetNext(obj InkObj) {
	panic(errors.Errorf("choices can not set next: %v", obj))
}

func (c *Choices) Next() InkObj {
	// panic(errors.New("choices can not go next"))
	return nil
}

func (c *Choices) Options() []*Option {
	return c.options
}

func (c *Choices) Select(idx int) *Option {
	if idx > (len(c.options)-1) || idx < 0 {
		return nil
	}

	c.story.current = c.options[idx]
	return c.options[idx]
}

func (c *Choices) Nesting() int {
	return c.nesting
}

// Option node of the choices
type Option struct {
	Inline
}
