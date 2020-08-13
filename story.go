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
		if opt := c.choose(idx); opt != nil {
			s.current = opt
			s.objCount[opt.Path()]++
			return opt
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

// findContainer of the current inkObj
func (s *Story) findContainer(obj InkObj) (*Knot, *Stitch) {
	for obj != nil {
		if st, ok := obj.(*Stitch); ok {
			return st.knot, st
		} else if k, ok := obj.(*Knot); ok {
			return k, nil
		}
		obj = obj.Parent()
	}

	return nil, nil
}

// findDivertCount in the given path
func (s *Story) findDivertCount(path string, obj InkObj) int {
	if res := s.findDivert(path, obj); res != nil {
		if count, ok := s.objCount[res.Path()]; ok {
			return count
		}
	}
	return 0
}

// findKnot of the story by name
func (s *Story) findKnot(name string) *Knot {
	if k, ok := s.objMap[name]; ok {
		if knot, b := k.(*Knot); b {
			return knot
		}
	}

	return nil
}

// findDivert in the given path
func (s *Story) findDivert(path string, obj InkObj) InkObj {
	sp := strings.Split(path, ".")
	knot, st := s.findContainer(obj)

	switch len(sp) {
	case 1: // local label || local stitch || story's knot
		// local label
		if knot != nil && st != nil {
			p := knot.name + split + st.name + split + path
			if s.objMap[p] != nil {
				return s.objMap[p]
			}
		}
		// find local stitch
		if knot != nil && knot.findStitch(path) != nil {
			return knot.findStitch(path)
		}
		// global knot
		if s.findKnot(path) != nil {
			return s.findKnot(path)
		}
	case 2: // local stitch.label || knot.stitch
		if knot != nil {
			p := regReplaceDot.ReplaceAllString(path, split+"$1")
			p = knot.name + split + p
			if s.objMap[p] != nil {
				return s.objMap[p]
			}
		}
		if k := s.findKnot(sp[0]); k != nil {
			return k.findStitch(sp[1])
		}
	default: // could be - knot.stitch.label
		p := regReplaceDot.ReplaceAllString(path, split+"$1")
		// fmt.Println(path)
		return s.objMap[p]
	}
	return nil
}

// parseLine input string into story's content
func (s *Story) parseLine(input string) error {
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

// ErrNotMatch the regexp error
var ErrNotMatch error = errors.New("RegExp Not Match")

// ParseFunc of the story
type ParseFunc func(s *Story, input string) error

var parsers []ParseFunc

const split string = "__"

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
