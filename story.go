package goink

import (
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// Story of the ink
type Story struct {
	start InkObj
	c     InkObj // current obj
	knots []*knot

	paths map[string]InkObj
	vars  map[string]int

	mux sync.Mutex
}

// InkObj is the basic element of a story
type InkObj interface {
	Story() *Story
	Parent() InkObj

	Next() InkObj
	SetNext(obj InkObj)

	Path() string
}

// current content of the story
func (s *Story) current() InkObj {
	return s.c
}

// reset the current content to start
func (s *Story) reset() {
	s.c = s.start
	s.vars = make(map[string]int)
}

// next content of the story
func (s *Story) next() InkObj {
	if n := s.c.Next(); n != nil {
		s.c = n
		s.vars[s.c.Path()]++

		return n
	}
	return nil
}

// choose the option of the current choices
func (s *Story) choose(idx int) *opt {
	if c, ok := s.c.(*options); ok {
		if opt := c.choose(idx); opt != nil {
			s.c = opt
			s.vars[opt.Path()]++
			return opt
		}
	}

	return nil
}

// Save current state of the story
func (s *Story) Save() *State {
	return NewState(s)
}

// Continue the story with the given vars
func (s *Story) Continue(state *State) (sec *Section, err error) {
	defer s.mux.Unlock()
	s.mux.Lock()

	sec = &Section{}

	for s.next() != nil {
		// t.Log(s.current.Path())
		switch s.c.(type) {
		case *line:
			sec.text += s.c.(*line).render()
			sec.tags = append(sec.tags, s.c.(*line).tags...)
			//t.Log(s.c.(*line).Render())
		case *opt:
			sec.text += s.c.(*opt).render(false)
			sec.tags = append(sec.tags, s.c.(*line).tags...)
			// t.Log(s.c.(*opt).render(false))
		case *gather:
			// t.Log(s.c.(*gather).Render())
		case *options:
			// t.Log("*", o.render(true))
		default: //knot or stitch - do nothing
		}
	}

	return
}

// Load previous state into story
func (s *Story) Load(state *State) error {
	if obj, ok := s.paths[state.path]; ok {
		s.c = obj
	} else {
		return errors.Errorf("cannot find the obj: %s", state.path)
	}

	// copy all non-zero count into story's count
	for k, v := range state.count {
		s.vars[k] = v
	}

	return nil
}

// find container of the current inkObj
func (s *Story) findContainer(obj InkObj) (*knot, *Stitch) {
	for obj != nil {
		if st, ok := obj.(*Stitch); ok {
			return st.k, st
		} else if k, ok := obj.(*knot); ok {
			return k, nil
		}
		obj = obj.Parent()
	}

	return nil, nil
}

// find divert count in the given path
func (s *Story) findDivertCount(path string, obj InkObj) int {
	if res := s.findDivert(path, obj); res != nil {
		if count, ok := s.vars[res.Path()]; ok {
			return count
		}
	}
	return 0
}

// find knot of the story by name
func (s *Story) findKnot(name string) *knot {
	if k, ok := s.paths[name]; ok {
		if kn, b := k.(*knot); b {
			return kn
		}
	}

	return nil
}

// find divert in the given path
func (s *Story) findDivert(path string, obj InkObj) InkObj {
	sp := strings.Split(path, ".")
	kn, st := s.findContainer(obj)

	switch len(sp) {
	case 1: // local label || local stitch || story's knot
		// local label
		if kn != nil && st != nil {
			p := kn.name + SPLIT + st.name + SPLIT + path
			if s.paths[p] != nil {
				return s.paths[p]
			}
		}
		// find local stitch
		if kn != nil && kn.findStitch(path) != nil {
			return kn.findStitch(path)
		}
		// global knot
		if s.findKnot(path) != nil {
			return s.findKnot(path)
		}
	case 2: // local stitch.label || knot.stitch
		if kn != nil {
			p := regReplaceDot.ReplaceAllString(path, SPLIT+"$1")
			p = kn.name + SPLIT + p
			if s.paths[p] != nil {
				return s.paths[p]
			}
		}
		if k := s.findKnot(sp[0]); k != nil {
			return k.findStitch(sp[1])
		}
	default: // could be - knot.stitch.label
		p := regReplaceDot.ReplaceAllString(path, SPLIT+"$1")
		// fmt.Println(path)
		return s.paths[p]
	}
	return nil
}

// parse single line input into story's content
func (s *Story) parseLine(input string) error {
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

// SPLIT of the path
const SPLIT string = "__"

// Parse lines of ink file into new story
func Parse(input string) (*Story, error) {
	contents := strings.Split(input, "\n")

	s := NewStory()

	for _, line := range contents {
		if err := s.parseLine(line); err != nil {
			return nil, err
		}
	}

	s.reset()
	return s, nil
}

// NewStory create an empty story instance
func NewStory() *Story {
	// Inline always be the last parser
	parsers = append(parsers, readKnot, readStitch, readOption, readGather, readLine)

	start := &line{raw: "[start]", path: "r"}

	story := &Story{start: start}
	story.c = story.start

	story.paths = make(map[string]InkObj)
	story.vars = make(map[string]int)

	return story
}
