package goink

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var choicesReg = regexp.MustCompile(`(^(\+\s*)+|^(\*\s*)+)(.+)`)

// NewOption parse and insert a new option into story
func NewOption(s *Story, input string) error {
	res := choicesReg.FindStringSubmatch(input)

	if res != nil {
		nesting := len(strings.Join(strings.Fields(res[1]), ""))

		i, err := CreateNewInline(res[4])
		if err != nil {
			return err
		}

		o := &Option{Inline: i}
		o.story = s

		// once-only option
		if len(res[2]) > 0 {
			o.sticky = true
		} else {
			o.sticky = false
		}

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

			choices.path = s.current.Path() + ".c"
			s.objMap[choices.path] = choices

			s.current.SetNext(choices)
		}

		o.path = choices.path + "." + strconv.Itoa(len(choices.options))
		o.parseCondition()

		choices.options = append(choices.options, o)
		s.objMap[o.path] = o

		// s.current.SetNext(o)
		o.parent = choices
		s.current = o

		// condition parse

		return nil
	}

	return ErrNotMatch
}

// Choices of the story
type Choices struct {
	story   *Story
	parent  InkObj
	options []*Option
	path    string

	gather  *Gather
	nesting int
}

// Story of the choices
func (c *Choices) Story() *Story {
	return c.story
}

// Path of the choices
func (c *Choices) Path() string {
	return c.path
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
func (c *Choices) Options() (os []*Option) {
	for _, opt := range c.options {
		if opt.sticky {
			os = append(os, opt)
		} else if count, ok := c.story.objCount[opt.Path()]; !ok || count == 0 {
			os = append(os, opt)
		}
	}
	return os
}

// Select the option of the choices by index
func (c *Choices) Select(idx int) *Option {
	// filtered options
	opts := c.Options()

	if idx >= len(opts) || idx < 0 {
		return nil
	}

	opt := opts[idx]
	return opt
}

// Nesting of the choices
func (c *Choices) Nesting() int {
	return c.nesting
}

// Option node of the choices
type Option struct {
	*Inline

	sticky    bool
	condition *Condition
}

var (
	exprReg       = regexp.MustCompile(`^\{(.+)\}(.*)`)
	supressingReg = regexp.MustCompile(`(^.*)\[(.*)\](.*$)`)
)

// Render option text with supressing
func (o *Option) Render(supressing bool) string {
	res := supressingReg.FindStringSubmatch(o.text)
	if res != nil {
		before := res[1]
		middle := res[2]
		after := res[3]

		if supressing {
			return before + middle
		}
		return before + after
	}
	return o.text
}

func (o *Option) parseCondition() {
	if res := exprReg.FindStringSubmatch(o.text); res != nil {
		o.condition = NewCondition(strings.TrimSpace(res[1]))
		o.text = res[2]
	}
}
