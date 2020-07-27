package internal

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Inline block of the story
type Inline interface {
	Story() *Story
	Prev() Inline
}

// Next block of the story
type Next interface {
	SetNext(prev Inline)
	Next() Inline
}

// Story of the ink
type Story struct {
	start Inline // start line of the story
	// end   Inline

	current Inline //current line of the story
}

// Reset the story
func (s *Story) Reset() {
	s.current = s.start
}

// Next content of the story
func (s *Story) Next() (Inline, error) {
	if next, ok := s.current.(Next); ok {
		s.current = next.Next()
		return s.current, nil
	}
	return nil, errors.New("cannot go next")
}

// Select the choice
func (s *Story) Select(idx int) (Inline, error) {
	if choices, ok := s.current.(*Choices); ok {
		return choices.Select(idx)
	}
	return nil, errors.New("cannot select")
}

var (
	choiceReg = regexp.MustCompile(`((^\++)|(^\*+))\s(.+)`)
)

// Parse input string into contents
func Parse(s *Story, input string) error {
	// trim spaces and skip empty lines
	input = strings.TrimRight(strings.TrimSpace(input), "\r\n")
	if len(input) == 0 {
		return nil
	}

	if s.current == nil {
		s.current = s.start
	}

	next, canNext := s.current.(Next)

	// * choices
	result := choiceReg.FindStringSubmatch(input)
	if result != nil {
		nesting := len(result[2]) + len(result[3])
		c := &PlainText{s: s, raw: result[4]}
		choices := findChoices(s, nesting)
		if choices == nil {
			choices = &Choices{s: s, p: s.current, nesting: nesting}
			next.SetNext(choices)
		}

		c.p = choices
		choices.selections = append(choices.selections, c)

		s.current = c
		return nil
	}

	// plain text
	if canNext {
		return errors.Errorf("current block cannot continue: %s", input)
	}

	p := &PlainText{s: s, p: s.current, raw: input}

	next.SetNext(p)
	s.current = p
	return nil
}

func findChoices(s *Story, nesting int) *Choices {
	inline := s.current

	for inline != nil {
		if choices, ok := inline.(*Choices); ok {
			if choices.nesting < nesting {
				return nil
			} else if choices.nesting == nesting {
				s.current = choices
				return choices
			}
		}

		inline = inline.Prev()
	}

	return nil
}

// NewStory of the Ink
func NewStory() *Story {
	return &Story{}
}

// PlainText of the story
type PlainText struct {
	s *Story // story
	p Inline // prev
	n Inline // next

	raw string
}

// Story of the content
func (p *PlainText) Story() *Story {
	return p.s
}

// Next content
func (p *PlainText) Next() Inline {
	return p.n
}

// SetNext content
func (p *PlainText) SetNext(prev Inline) {
	p.n = prev
}

// Prev content
func (p *PlainText) Prev() Inline {
	return p.p
}

// Choices of the story
type Choices struct {
	s *Story // story
	p Inline // prev

	nesting    int
	selections []Inline
}

// Story of the content
func (c *Choices) Story() *Story {
	return c.s
}

// Prev content
func (c *Choices) Prev() Inline {
	return c.p
}

// Select  content
func (c *Choices) Select(idx int) (Inline, error) {
	idx = idx - 1
	if idx < 0 || idx >= len(c.selections) {
		return nil, errors.Errorf("Invalid selection index: %d", idx)
	}

	c.s.current = c.selections[idx]
	return c.s.current, nil
}
