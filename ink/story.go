package ink

import (
	"sync"

	"github.com/pkg/errors"
)

var (
	parsers []ParseFunc
)

const (
	// PathSplit sign of the path
	PathSplit string = "__"
)

// Node is the basic element of a story
type Node interface {
	Story() *Story
	Parent() Node

	Path() string
}

type base struct {
	story  *Story
	parent Node
	path   string
}

func (b *base) Story() *Story {
	return b.story
}

func (b *base) Parent() Node {
	return b.parent
}

func (b *base) Path() string {
	return b.path
}

// End of story
type End interface {
	End() (text string, tags []string)
}

// Choices content - which has one/more option(s)
type Choices interface {
	Select(idx int) (Node, error)
	List() (text []string, tags [][]string)
}

// CanNext content - which can go next
type CanNext interface {
	Next() Node
	SetNext(node Node)
	Render() (text string, tags []string)
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

// Resume the story
func (s *Story) Resume(ctx *Context) (sec *Section, err error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if err := s.load(ctx); err != nil {
		return nil, err
	}
	return s.resume()
}

// Pick the option
func (s *Story) Pick(ctx *Context, idx int) (sec *Section, err error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if err := s.load(ctx); err != nil {
		return nil, err
	}
	return s.pick(idx)
}

// resume the story
func (s *Story) resume() (sec *Section, err error) {
	var nodes Nodes
loop:
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
			break loop
		case Choices:
			break loop
		case CanNext:
			s.current = s.current.(CanNext).Next()
		default:
			return nil, errors.Errorf("current node is not recgonized: %s", s.current.Path())
		}
	}

	return nodes.NewSection(), nil
}

// pick one of the current choices' option,
// and resume
func (s *Story) pick(idx int) (sec *Section, err error) {
	c, err := s.choose(idx)
	if err != nil {
		return nil, err
	}
	s.current = c
	return s.resume()
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

// load from context
func (s *Story) load(ctx *Context) error {
	n, ok := s.paths[ctx.current]
	if !ok {
		return errors.Errorf("current path [%s] is not existed", ctx.current)
	}

	s.current = n
	s.vars = ctx.vars
	return nil
}

// Context of the story
type Context struct {
	current string
	vars    map[string]interface{}
}

// Current path of the story
func (c *Context) Current() string {
	return c.current
}

// Vars of the story
func (c *Context) Vars() map[string]interface{} {
	return c.vars
}

// Section is rendered result of current story
type Section struct {
	text string
	tags []string

	opts     []string
	optsTags [][]string

	end bool
}

func (s *Section) add(text string, tags []string) {
	if len(text) > 0 {
		s.text = s.text + "\n" + text
	}

	s.tags = append(s.tags, tags...)
}

// Nodes list
type Nodes []Node

// NewSection creates the rendered result of the nodes
func (n Nodes) NewSection() *Section {
	sec := &Section{}
	for _, node := range n {
		switch node.(type) {
		case End:
			sec.end = true
			sec.add(node.(End).End())
		case Choices:
			opts, optsTags := node.(Choices).List()

			sec.opts = opts
			sec.optsTags = optsTags
		case CanNext:
			sec.add(node.(CanNext).Render())
		}
	}
	return sec
}

// ParseFunc of the story
type ParseFunc func(s *Story, input string) error

// Default story
func Default() *Story {
	// parsers = append(parsers, readLine)
	story := &Story{}
	return story
}
