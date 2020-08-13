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

		i, err := createNewInline(res[4])
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

			choices.path = s.current.Path() + SPLIT + "c"
			s.objMap[choices.path] = choices
			s.current.SetNext(choices)
		}

		o.path = choices.path + SPLIT + strconv.Itoa(len(choices.opts))

		choices.opts = append(choices.opts, o)
		s.objMap[o.path] = o

		// s.current.SetNext(o)
		o.parent = choices
		s.current = o

		// post parsing process
		if err := o.parseCondition(); err != nil {
			return err
		}
		if err := o.parseLabel(); err != nil {
			return err
		}
		return nil
	}

	return ErrNotMatch
}

// Choices of the story
type Choices struct {
	story  *Story
	parent InkObj
	opts   []*Option
	path   string

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

// options of the choices
func (c *Choices) options() (os []*Option) {
	for _, opt := range c.opts {
		// condition test
		if opt.condition != nil {
			b, err := opt.condition.Bool(c.story.objCount)
			if err != nil {
				panic(err)
			}

			// will not display, when condition test is false
			// no matter sticky or not
			if !b {
				continue
			}
		}

		// sticky or once-only
		if opt.sticky {
			os = append(os, opt)
		} else if count, ok := c.story.objCount[opt.Path()]; !ok || count == 0 {
			os = append(os, opt)
		}
	}
	return os
}

// choose the option of the choices by index
func (c *Choices) choose(idx int) *Option {
	// filtered options
	opts := c.options()

	if idx >= len(opts) || idx < 0 {
		return nil
	}

	opt := opts[idx]
	return opt
}

// Option node of the choices
type Option struct {
	*Inline

	sticky    bool
	condition *Condition
}

var (
	exprReg       = regexp.MustCompile(`^\s*\{(.+)\}(.*)`)
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

func (o *Option) parseCondition() error {
	if res := exprReg.FindStringSubmatch(o.text); res != nil {
		if c, err := NewCondition(strings.TrimSpace(res[1])); err == nil {
			o.condition = c
			o.text = res[2]
		} else {
			return err
		}
	}

	return nil
}

func (o *Option) parseLabel() error {
	if res := lableReg.FindStringSubmatch(o.text); res != nil {
		label := strings.TrimSpace(res[1])
		if len(label) > 0 {
			if knot, stitch := o.story.findContainer(o); stitch != nil {
				label = stitch.Path() + SPLIT + label
			} else if knot != nil {
				label = knot.Path() + SPLIT + label
			}

			if _, ok := o.story.objMap[label]; ok {
				return errors.Errorf("duplicated label: %s", label)
			}
			o.story.objMap[label] = o
			o.path = label
		}
		o.text = res[2]
	}

	return nil
}
