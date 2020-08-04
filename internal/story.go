package internal

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Story of the ink
type Story struct {
	start   InkObj
	current InkObj
}

func (s *Story) Current() InkObj {
	return s.current
}

var (
	choicesReg = regexp.MustCompile(`(^(\+\s*)+|^(\*\s*)+)(.+)`)
	gatherReg  = regexp.MustCompile(`^((-\s*)+)([^>].+)`)
)

func NewStory() *Story {
	start := &Inline{raw: "[start]"}
	story := &Story{start: start}
	story.current = story.start
	return story
}

func (s *Story) Reset() {
	s.current = s.start
}

func (s *Story) Next() InkObj {
	if next := s.current.Next(); next != nil {
		s.current = next
		return next
	}
	return nil
}

func (s *Story) Parse(input string) error {
	// trim spaces and skip empty lines
	input = strings.TrimRight(strings.TrimSpace(input), "\r\n")
	if len(input) == 0 {
		return nil
	}

	// Choices
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

	// Gather
	res = gatherReg.FindStringSubmatch(input)
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
			g.parent = choices
			choices.gather = g
			s.current = g

			return nil
		}

		return errors.Errorf("wrong nesting of the gather %s", input)
	}

	// Inline
	i := &Inline{story: s, parent: s.current, raw: input}
	s.current.SetNext(i)
	s.current = i
	return nil
}

// Basic Node of the story
type InkObj interface {
	Story() *Story
	Parent() InkObj

	Next() InkObj
	SetNext(obj InkObj)
}

// Inline node of tye story
type Inline struct {
	story  *Story
	parent InkObj
	next   InkObj

	raw string
}

func (i *Inline) Render() string {
	return i.raw
}

func (i *Inline) Parent() InkObj {
	return i.parent
}

func (i *Inline) SetNext(obj InkObj) {
	i.next = obj
}

func (i *Inline) Next() InkObj {
	if i.next != nil {
		return i.next
	}

	obj := i.parent

	for obj != nil {
		if c, ok := obj.(*Choices); ok {
			if c.gather != nil {
				return c.gather
			}
		}

		if _, ok := obj.(*Gather); ok {
			obj = obj.Parent()
		}

		obj = obj.Parent()
	}

	return nil
}

func (i *Inline) Story() *Story {
	return i.story
}

// Option node of the choices
type Option struct {
	Inline
}

// Gather node of the choices
type Gather struct {
	Inline
	nesting int
}

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

		if _, ok := obj.(*Gather); ok {
			obj = obj.Parent()
		}

		obj = obj.Parent()
	}

	return nil
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
