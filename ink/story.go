package ink

import (
	"sync"

	"github.com/pkg/errors"
)

// Node is the basic element of a story
type Node interface {
	Story() *Story
	Parent() Node

	Path() string
}

type End interface {
	End() (text string, tags []string)
}

// Choices content - which has one/more option(s)
type Choices interface {
	Select(idx int) (Node, error)
	List() (text []string, tags [][]string)
}

// Flow content - which can go next
type CanNext interface {
	Next() Node
	SetNext(obj Node)

	Render() (text string, tags []string)
}

type Nodes []Node

type Context struct {
	current string
	vars    map[string]interface{}
}

// Story of the ink
type Story struct {
	current Node
	vars    map[string]interface{}

	start Node
	end   Node

	paths map[string]Node

	mux sync.Mutex
}

// Continue the story
func (s *Story) Continue() (nodes Nodes, err error) {
	for {
		if s.current == nil {
			return nil, errors.New("current node is nil")
		}

		nodes = append(nodes, s.current)
		if err := s.visit(s.current.Path()); err != nil {
			return nil, err
		}

		switch s.current.(type) {
		case End:
			return
		case Choices:
			return
		case CanNext:
			s.current = s.current.(CanNext).Next()
		}
	}
}

// Choose one of the current choices' option,
// and Continue
func (s *Story) Choose(idx int) (nodes Nodes, err error) {
	if c, err := s.choose(idx); err != nil {
		return nil, err
	} else {
		s.current = c
		return s.Continue()
	}
}

func (s *Story) isNext() CanNext {
	if current, ok := s.current.(CanNext); ok {
		return current
	}

	return nil
}

func (s *Story) isChoices() Choices {
	if current, ok := s.current.(Choices); ok {
		return current
	}

	return nil
}

// add visit count to the given path
func (s *Story) visit(path string) error {
	if v, ok := s.vars[path]; ok {
		if n, ok := v.(int); ok {
			n++
			s.vars[path] = n
		} else {
			return errors.Errorf("'%s' is not [int]", path)
		}
	}

	s.vars[path] = 1
	return nil
}

// choose from current choices by index
func (s *Story) choose(idx int) (Node, error) {
	if c := s.isChoices(); c != nil {
		return c.Select(idx)
	}

	return nil, errors.Errorf("%s is not [Choices]", s.current.Path())
}
