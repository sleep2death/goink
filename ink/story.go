package ink

import (
	"strings"
	"sync"

	"github.com/pkg/errors"
)

var (
	ErrNotMatch error = errors.New("RegExp Not Match")
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
	Nesting() int
}

// CanNext content - which can go next
type CanNext interface {
	Next() (Node, error)
	SetNext(node Node)
	Render() (text string, tags []string)
}

// Story of the ink
type Story struct {
	current Node
	vars    map[string]interface{}

	start Node
	end   Node

	paths   map[string]Node
	parsers []ParseFunc

	mux sync.Mutex
}

// Resume the story
func (s *Story) Resume(ctx *Context) (sec *Section, err error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if err := s.load(ctx); err != nil {
		return nil, err
	}

	if sec, err = s.resume(); err != nil {
		return nil, err
	}

	// update ctx
	*ctx = s.save()
	return
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

		switch node := s.current.(type) {
		case End:
			break loop
		case Choices:
			break loop
		case CanNext:
			n, err := node.Next()
			if err != nil {
				return nil, err
			}
			s.current = n
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

func (s *Story) next() CanNext {
	if current, ok := s.current.(CanNext); ok {
		return current
	}

	return nil
}

func (s *Story) choices() Choices {
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
	if c := s.choices(); c != nil {
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

func (s *Story) save() Context {
	return Context{current: s.current.Path(), vars: copyMap(s.vars)}
}

func copyMap(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{})
	for k, v := range m {
		vm, ok := v.(map[string]interface{})
		if ok {
			cp[k] = copyMap(vm)
		} else {
			cp[k] = v
		}
	}

	return cp
}

func (s *Story) divert(path string, from Node) Node {
	sp := strings.Split(path, ".")
	if len(sp) == 1 {
		return s.paths[path]
	}
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
		switch node := node.(type) {
		case End:
			sec.end = true
			sec.add(node.End())
		case Choices:
			opts, optsTags := node.List()

			sec.opts = opts
			sec.optsTags = optsTags
		case CanNext:
			sec.add(node.Render())
		}
	}
	return sec
}

type start struct {
	*base
	next Node
}

func (s *start) SetNext(n Node) {
	s.next = n
}

func (s *start) Next() (Node, error) {
	if s.next == nil {
		return nil, errors.New("current node can not go next: [start]")
	}

	return s.next, nil
}

func (s *start) Render() (text string, tags []string) {
	text = "[start]"
	tags = append(tags, "START")
	return
}

type end struct {
	*base
}

func (e *end) End() (text string, tags []string) {
	text = "[end]"
	tags = append(tags, "END")
	return
}

// Default story
func Default() *Story {
	parsers := []ParseFunc{readLine}

	s := &start{base: &base{path: "start"}}
	e := &end{base: &base{path: "end"}}
	s.SetNext(e)

	story := &Story{start: s, end: e, parsers: parsers}

	s.story = story
	e.story = story

	story.paths = make(map[string]Node)
	story.vars = make(map[string]interface{})

	story.paths["start"] = s
	story.paths["end"] = e

	story.current = s
	return story
}

// ParseFunc of the story
type ParseFunc func(s *Story, input string) error

func (s *Story) Parse(input string) error {
	contents := strings.Split(input, "\n")
	for _, line := range contents {
		// trim spaces and skip empty lines
		l := strings.TrimRight(strings.TrimSpace(line), "\r\n")
		if len(l) == 0 {
			continue
		}

		// passing raw input into parsers
		for _, parser := range s.parsers {
			if err := parser(s, l); err != nil {
				if err != ErrNotMatch {
					return err
				}
			} else {
				continue
			}
		}
	}

	return nil
}
