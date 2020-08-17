package goink

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var optsReg = regexp.MustCompile(`(^(\+\s*)+|^(\*\s*)+)(.+)`)

// readOption parse and insert a new option into story
func readOption(s *Story, input string) error {
	res := optsReg.FindStringSubmatch(input)

	if res != nil {
		nesting := len(strings.Join(strings.Fields(res[1]), ""))

		i, err := newLine(res[4])
		if err != nil {
			return err
		}

		o := &opt{line: i}
		o.story = s

		// once-only option
		if len(res[2]) > 0 {
			o.sticky = true
		} else {
			o.sticky = false
		}

		obj := s.c

		var choices *options
		// var gather *Gather

		for obj != nil {
			if g, ok := obj.(*gather); ok {
				if t := nesting - g.nesting; t == 0 {
					break
				}
			}

			if c, ok := obj.(*options); ok {
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
			choices = &options{story: s, parent: s.c, nesting: nesting}

			choices.path = s.c.Path() + SPLIT + "c"
			s.paths[choices.path] = choices

			s.setNext(choices)
		}

		o.path = choices.path + SPLIT + strconv.Itoa(len(choices.opts))

		choices.opts = append(choices.opts, o)
		s.paths[o.path] = o

		// s.current.SetNext(o)
		o.parent = choices
		s.c = o

		// post parsing process
		if err := o.parseExprc(); err != nil {
			return err
		}
		if err := o.parseLabel(); err != nil {
			return err
		}
		return nil
	}

	return ErrNotMatch
}

// options of the story
type options struct {
	story  *Story
	parent Node
	opts   []*opt
	path   string

	gather  *gather
	nesting int
}

// Story of the choices
func (c *options) Story() *Story {
	return c.story
}

// Path of the choices
func (c *options) Path() string {
	return c.path
}

// Parent of the choices
func (c *options) Parent() Node {
	return c.parent
}

// options of the choices
func (c *options) list() (os []*opt) {
	for _, opt := range c.opts {
		// condition test
		if opt.condition != nil {
			b, err := opt.condition.Bool(c.story.vars)
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
		} else if count, ok := c.story.vars[opt.Path()]; !ok || count == 0 {
			os = append(os, opt)
		}
	}
	return os
}

// List all available options' content
func (c *options) List() (text []string, tags [][]string) {
	if opts := c.list(); len(opts) > 0 {
		for _, opt := range opts {
			str, tag := opt.List()
			text = append(text, str)
			tags = append(tags, tag)
		}

		return
	}

	panic(errors.Errorf("no option available: %s", c.Path()))
}

// choose the option of the choices by index
func (c *options) choose(idx int) *opt {
	// filtered options
	opts := c.list()

	if idx >= len(opts) || idx < 0 {
		return nil
	}

	opt := opts[idx]
	return opt
}

func (c *options) Select(idx int) (Node, error) {
	res := c.choose(idx)
	if res == nil {
		return nil, errors.Errorf("no option available [%s] at idx: %d", c.Path(), idx)
	}

	return c.choose(idx), nil
}

// Opt of the options
type opt struct {
	*line

	sticky    bool
	condition *exprc
}

var (
	exprReg       = regexp.MustCompile(`^\s*\{(.+)\}(.*)`)
	supressingReg = regexp.MustCompile(`(^.*)\[(.*)\](.*$)`)
)

// render option text with supressing
func (o *opt) render(supressing bool) string {
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

// Render option text with supressing = false
func (o *opt) Render() (str string, tags []string) {
	return o.render(false), o.tags
}

// List option text with supressing = true
func (o *opt) List() (str string, tags []string) {
	return o.render(true), o.tags
}

func (o *opt) parseExprc() error {
	if res := exprReg.FindStringSubmatch(o.text); res != nil {
		if c, err := newExprc(strings.TrimSpace(res[1])); err == nil {
			o.condition = c
			o.text = res[2]
		} else {
			return err
		}
	}

	return nil
}

func (o *opt) parseLabel() error {
	if res := lableReg.FindStringSubmatch(o.text); res != nil {
		label := strings.TrimSpace(res[1])
		if len(label) > 0 {
			if knot, stitch := o.story.findContainer(o); stitch != nil {
				label = stitch.Path() + SPLIT + label
			} else if knot != nil {
				label = knot.Path() + SPLIT + label
			}

			if _, ok := o.story.paths[label]; ok {
				return errors.Errorf("duplicated label: %s", label)
			}
			o.story.paths[label] = o
			o.path = label
		}
		o.text = res[2]
	}

	return nil
}
