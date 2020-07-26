package goink

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Story of the ink
type Story struct {
	start Node // start line of the story
	end   Node

	knots []*Knot

	current Node //current line of the story
}

// Reset the story
func (s *Story) Reset() {
	s.current = s.start
}

// Next content of the story
func (s *Story) Next() (Node, error) {
	if next, ok := s.current.(Next); ok {
		s.current = next.Next()
		if s.current != nil {
			return s.current, nil
		}
	}
	return nil, errors.New("cannot go next")
}

// Select the choice
func (s *Story) Select(idx int) (Node, error) {
	if choices, ok := s.current.(*Choices); ok {
		return choices.Select(idx)
	}
	return nil, errors.New("cannot select")
}

// FindKnot of the story by name
func (s *Story) FindKnot(name string) *Knot {
	for _, k := range s.knots {
		if k.name == name {
			return k
		}
	}
	return nil
}

var (
	choiceReg = regexp.MustCompile(`((^\++)|(^\*+))\s(.+)`)
	knotReg   = regexp.MustCompile(`(^\={1,})\s(\w+)`)
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

	if canNext {
		errors.Errorf("current block cannot continue: %s", input)
	}

	// == knot
	result := knotReg.FindStringSubmatch(input)
	if result != nil {
		k := &Knot{s: s, name: result[2]}
		s.knots = append(s.knots, k)
		s.current = k

		return nil
	}

	// * choices
	result = choiceReg.FindStringSubmatch(input)
	if result != nil {
		nesting := len(result[2]) + len(result[3])
		// c := &Inline{s: s, raw: result[4]}
		c := NewInline(result[4])
		c.s = s
		choices := findChoices(s, nesting)
		if choices == nil {
			choices = &Choices{s: s, p: s.current, nesting: nesting}
			next.SetNext(choices)
		}

		// add plain text of the choice into choices,
		// and make it the current node
		c.p = choices
		choices.selections = append(choices.selections, c)

		s.current = c
		return nil
	}

	// plain text
	p := NewInline(input)
	p.s = s
	p.p = s.current

	next.SetNext(p)
	s.current = p
	return nil
}

func findChoices(s *Story, nesting int) *Choices {
	inline := s.current

	for {
		if choices, ok := inline.(*Choices); ok {
			if choices.nesting < nesting {
				return nil
			} else if choices.nesting == nesting {
				s.current = choices
				return choices
			}
		}

		if p, ok := inline.(Prev); ok {
			inline = p.Prev()
		} else {
			return nil
		}
	}
}

// NewStory of the Ink
func NewStory() *Story {
	return &Story{}
}

// Knot of the story
type Knot struct {
	s *Story
	n Node

	name string // name of the knot
}

// Story of the knot
func (k *Knot) Story() *Story {
	return k.s
}

// Next content
func (k *Knot) Next() Node {
	return k.n
}

// SetNext content
func (k *Knot) SetNext(next Node) {
	k.n = next
}
