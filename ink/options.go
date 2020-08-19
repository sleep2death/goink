package ink

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	optsReg = regexp.MustCompile(`(^(\+\s*)+|^(\*\s*)+)(.+)`)
)

// readOption parse and insert a new option into story
func readOption(s *Story, input string) error {
	res := optsReg.FindStringSubmatch(input)

	if res != nil {
		// findout nesting num
		nesting := len(strings.Join(strings.Fields(res[1]), ""))

		// create new option
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

		// find parent options
		opts, err := s.findParentOptions(nesting)
		if err != nil {
			return err
		}

		if opts == nil {
			opts = &options{base: &base{story: s, parent: s.current}, nesting: nesting}

			opts.path = s.current.Path() + PathSplit + "c"
			s.paths[opts.path] = opts

			if n := s.next(); n != nil {
				n.SetNext(opts)
			} else {
				return errors.Errorf("current node can not go next: [%s]", s.current.Path())
			}
		}

		// if option didn't have label
		if o.path == "" {
			o.path = opts.path + PathSplit + strconv.Itoa(len(opts.opts))
			s.paths[o.path] = o
		}
		opts.opts = append(opts.opts, o)

		o.parent = opts
		s.current = o

		// condition exprc parsing
		if err := o.parseExprc(); err != nil {
			return err
		}

		// parsing label
		if err := i.parseLabel(); err != nil {
			return err
		}

		return nil
	}

	return errNotMatch
}

func (s *Story) findParentOptions(nesting int) (opts *options, err error) {
	node := s.current

	for node != nil {
		if c, ok := node.(*options); ok {
			if t := nesting - c.nesting; t >= 0 {
				if t > 1 {
					return nil, errors.Errorf("wrong nesting of the option: %s", c.Path())
				} else if t == 0 {
					opts = c
				}
				break
			}
		}

		node = node.Parent()
	}

	return
}

// options of the story
type options struct {
	*base
	opts []*opt

	gather  *gather
	nesting int
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
			str, tag := opt.list()
			text = append(text, str)
			tags = append(tags, tag)
		}

		return
	}

	panic(errors.Errorf("no option available: %s", c.Path()))
}

func (c *options) pick(idx int) *opt {
	// filtered options
	opts := c.list()

	if idx >= len(opts) || idx < 0 {
		return nil
	}

	opt := opts[idx]
	return opt
}

// Pick the option of the choices by index
func (c *options) Pick(idx int) (Node, error) {
	res := c.pick(idx)
	if res == nil {
		return nil, errors.Errorf("no option available [%s] at idx: %d", c.Path(), idx)
	}

	return res, nil
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

// Render option text without supressing
func (o *opt) Render() (str string, tags []string) {
	return o.render(false), o.tags
}

// List option text with supressing = true
func (o *opt) list() (str string, tags []string) {
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
