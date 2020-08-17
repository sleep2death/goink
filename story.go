package goink

import (
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// Story of the ink
type Story struct {
	start Node
	end   Node

	c     Node // current obj
	knots []*knot

	paths map[string]Node
	vars  map[string]int

	mux sync.Mutex
}

// Node is the basic element of a story
type Node interface {
	Story() *Story
	Parent() Node

	Path() string
}

// Flow content - which can go next
type Flow interface {
	Next() Node
	SetNext(obj Node)

	Render() (text string, tags []string)
}

type End interface {
	End() (text string, tags []string)
}

// Choices content - which has one/more option(s)
type Choices interface {
	List() (text []string, tags [][]string)
	Select(idx int) (Node, error)
}

// current content of the story
func (s *Story) current() Node {
	return s.c
}

// reset the current content to start
func (s *Story) reset() {
	s.c = s.start
	s.vars = make(map[string]int)
}

// next content of the story
func (s *Story) next() Node {
	if current := s.canNext(); current != nil {
		if n := current.Next(); n != nil {
			s.c = n
			s.vars[s.c.Path()]++

			return n
		}
	}

	return nil
	// panic(errors.Errorf("current obj can not go next: %v", s.c))
}

// next content of the story
func (s *Story) canNext() Flow {
	if current, ok := s.c.(Flow); ok {
		return current
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
	return NewState(s, false)
}

// GoOn the story with the given state
func (s *Story) GoOn(state *State) (update *State, sec *Section, err error) {
	defer s.mux.Unlock()
	s.mux.Lock()

	if err := s.Load(state); err != nil {
		return nil, nil, err
	}

	sec = &Section{}
	s.goOn(sec)
	update = NewState(s, false)
	return
}

// Select the option with the given state
func (s *Story) Select(state *State, idx int) (update *State, sec *Section, err error) {
	defer s.mux.Unlock()
	s.mux.Lock()

	if err := s.Load(state); err != nil {
		return nil, nil, err
	}

	opts, ok := s.c.(*options)
	if !ok {
		return nil, nil, errors.Errorf("current node is not options: %s", s.c.Path())
	}

	node, err := opts.Select(idx)
	if err != nil {
		return nil, nil, err
	}

	s.c = node
	sec = &Section{}
	s.goOn(sec)

	update = NewState(s, false)
	return
}

func (s *Story) goOn(sec *Section) {
loop:
	for s.c != nil {
		// t.Log(s.current.Path())
		switch s.c.(type) {
		case End:
			str, tag := s.c.(End).End()
			if len(str) > 0 {
				sec.text = sec.text + "\n" + str
			}
			sec.tags = append(sec.tags, tag...)
			sec.end = true
			break loop
		case Choices:
			opts, tags := s.c.(Choices).List()
			sec.opts = opts
			sec.optsTags = tags
			break loop
		case Flow:
			str, tag := s.c.(Flow).Render()
			if len(str) > 0 {
				sec.text = sec.text + "\n" + str
			}
			sec.tags = append(sec.tags, tag...)
			s.vars[s.c.Path()]++
			s.c = s.c.(Flow).Next()
			//t.Log(s.c.(*line).Render())
		default: //end of story - do nothing
			panic(errors.Errorf("invalid node: %s", s.c.Path()))
		}
	}
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
func (s *Story) findContainer(obj Node) (*knot, *Stitch) {
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
func (s *Story) findDivertCount(path string, obj Node) int {
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
func (s *Story) findDivert(path string, obj Node) Node {
	sp := strings.Split(path, ".")
	kn, st := s.findContainer(obj)

	switch len(sp) {
	case 1: // local label || local stitch || story's knot
		if strings.ToLower(path) == "end" {
			return s.end
		}
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

func (s *Story) setNext(obj Node) {
	if current := s.canNext(); current != nil {
		current.SetNext(obj)
		s.c = obj
	} else {
		panic(errors.Errorf("current obj can not set next: %v", s.c))
	}
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

	story.paths = make(map[string]Node)
	story.vars = make(map[string]int)

	story.paths["r"] = start

	end := &end{&line{raw: "[end]", path: "end"}}
	story.end = end

	return story
}
