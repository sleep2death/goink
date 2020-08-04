package goink

import (
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

	for _, parser := range parsers {
		if err := parser(s, input); err != nil {
			if err != NotMatch {
				return err
			}
		} else {
			return nil
		}
	}

	return nil
}

var NotMatch = errors.New("RegExp Not Match")

type ParseFunc func(s *Story, input string) error

var parsers []ParseFunc

func NewStory() *Story {
	// Inline always be the last parser
	parsers = append(parsers, NewOption, NewGather, NewInline)

	start := &Inline{raw: "[start]"}
	story := &Story{start: start}
	story.current = story.start

	return story
}
