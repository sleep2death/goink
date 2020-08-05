package goink

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var choicesReg = regexp.MustCompile(`(^(\+\s*)+|^(\*\s*)+)(.+)`)

// NewOption parse and insert a new option into story
func NewOption(s *Story, input string) error {
	res := choicesReg.FindStringSubmatch(input)

	if res != nil {
		nesting := len(strings.Join(strings.Fields(res[1]), ""))

		i, err := CreateNewline(res[3])
		if err != nil {
			return err
		}

		o := &Option{Inline: i}
		o.story = s

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

	return ErrNotMatch
}

// Choices of the story
type Choices struct {
	story   *Story
	parent  InkObj
	options []*Option

	gather  *Gather
	nesting int
}

// Story of the choices
func (c *Choices) Story() *Story {
	return c.story
}

// Parent of the choices
func (c *Choices) Parent() InkObj {
	return c.parent
}

// SetNext of the choices should fire panic
func (c *Choices) SetNext(obj InkObj) {
	panic(errors.Errorf("choices can not set next: %v", obj))
}

// Next content of the choices should be nil
func (c *Choices) Next() InkObj {
	// panic(errors.New("choices can not go next"))
	return nil
}

// Options of the choices
func (c *Choices) Options() []*Option {
	return c.options
}

// Select the option of the choices by index
func (c *Choices) Select(idx int) *Option {
	if idx > (len(c.options)-1) || idx < 0 {
		return nil
	}

	c.story.current = c.options[idx]
	return c.options[idx]
}

// Nesting of the choices
func (c *Choices) Nesting() int {
	return c.nesting
}

// Option node of the choices
type Option struct {
	*Inline
}
