package goink

import (
	"strings"

	"github.com/pkg/errors"
)

// Story of the ink
type Story struct {
	start   InkObj
	current InkObj
	ln      int
	knots   []*Knot

	objMap   map[string]InkObj
	objCount map[string]int
	vars     map[string]interface{}
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
		s.objCount[s.current.Path()]++

		return next
	}
	return nil
}

// Select the option of the current choices
func (s *Story) Select(idx int) *Option {
	if c, ok := s.current.(*Choices); ok {
		if opt := c.Select(idx); opt != nil {
			s.current = opt
			s.objCount[opt.Path()]++
			return opt
		}
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

	s.ln++
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

// Save current state of the story
func (s *Story) Save() *State {
	return NewState(s)
}

// Load previous state into story
func (s *Story) Load(state *State) error {
	if obj, ok := s.objMap[state.path]; ok {
		s.current = obj
	} else {
		return errors.Errorf("cannot find the obj: %s", state.path)
	}

	// copy all non-zero count into story's count
	for k, v := range state.count {
		s.objCount[k] = v
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

	start := &Inline{raw: "[start]", path: "r"}

	story := &Story{start: start}
	story.current = story.start

	story.objMap = make(map[string]InkObj)
	story.objCount = make(map[string]int)
	story.vars = make(map[string]interface{})

	return story
}
