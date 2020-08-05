package goink

import (
	"strings"

	"github.com/pkg/errors"
)

// Story of the ink
type Story struct {
	start   InkObj
	current InkObj

	knots []*Knot
}

// Current content of the story
func (s *Story) Current() InkObj {
	return s.current
}

// Reset the current content to start
func (s *Story) Reset() {
	s.current = s.start
}

// Next content of the story
func (s *Story) Next() InkObj {
	if next := s.current.Next(); next != nil {
		s.current = next
		return next
	}
	return nil
}

// Parse input string into story's content
func (s *Story) Parse(input string) error {
	// trim spaces and skip empty lines
	input = strings.TrimRight(strings.TrimSpace(input), "\r\n")
	if len(input) == 0 {
		return nil
	}

	for _, parser := range parsers {
		if err := parser(s, input); err != nil {
			if err != ErrNotMatch {
				return err
			}
		} else {
			return nil
		}
	}

	return nil
}

// ErrNotMatch the regexp error
var ErrNotMatch error = errors.New("RegExp Not Match")

// ParseFunc of the story
type ParseFunc func(s *Story, input string) error

var parsers []ParseFunc

// NewStory create an empty story instance
func NewStory() *Story {
	// Inline always be the last parser
	parsers = append(parsers, NewKnot, NewStitch, NewOption, NewGather, NewInline)

	start := &Inline{raw: "[start]"}
	story := &Story{start: start}
	story.current = story.start

	return story
}
