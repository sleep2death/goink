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
		return s.current, nil
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
		errors.Errorf("current block cannot continue: %s", input)
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
