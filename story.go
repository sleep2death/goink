package goink

import (
	"strings"
	"sync"

	"github.com/pkg/errors"
)

var (
	// PathSplit of the node's path
	PathSplit string = "__"

	errNotMatch error = errors.New("RegExp Not Match")
)

// Node is the basic element of a story
type Node interface {
	Story() *Story
	Parent() Node
	PostParsing() error

	Path() string
}

// embeding struct which implements Node
type base struct {
	story  *Story
	parent Node
	path   string
}

// Story of the node
func (b *base) Story() *Story {
	return b.story
}

// Parent of the node
func (b *base) Parent() Node {
	return b.parent
}

// Path of the node
func (b *base) Path() string {
	return b.path
}

// do some post parsing check
func (b *base) PostParsing() error {
	return nil
}

// End of story
type End interface {
	End() (text string, tags []string)
}

// Choices content - which has one/more option(s)
type Choices interface {
	Pick(idx int) (Node, error)
	List() (text []string, tags [][]string)
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

	knots []*knot

	id  string // story's unique name
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

	sec, err = s.pick(idx)
	// update ctx
	*ctx = s.save()
	return
}

func (s *Story) SetID(id string) {
	s.id = id
}

// resume the story
func (s *Story) resume() (sec *Section, err error) {
	var ns nodes
loop:
	for {
		if s.current == nil {
			return nil, errors.New("current node is nil")
		}

		ns = append(ns, s.current)
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

	return ns.section(), nil
}

// pick one of the current choices' option,
// and resume
func (s *Story) pick(idx int) (sec *Section, err error) {
	var opt Node
	if c := s.choices(); c != nil {
		if opt, err = c.Pick(idx); err != nil {
			return nil, err
		}
		s.current = opt
		return s.resume()
	}

	return nil, errors.Errorf("%s is not Choices", s.current.Path())
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
			return errors.Errorf("variable: <%s> is not type of int", path)
		}
	}

	s.vars[path] = 1
	return nil
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
	return Context{current: s.current.Path(), vars: copy(s.vars)}
}

func copy(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{})
	for k, v := range m {
		vm, ok := v.(map[string]interface{})
		if ok {
			cp[k] = copy(vm)
		} else {
			cp[k] = v
		}
	}

	return cp
}

// container of the current node
func (s *Story) container(node Node) (*knot, *stitch) {
	for node != nil {
		if st, ok := node.(*stitch); ok {
			return st.knot, st
		} else if k, ok := node.(*knot); ok {
			return k, nil
		}
		node = node.Parent()
	}

	return nil, nil
}

// find knot of the story by name
func (s *Story) knot(name string) *knot {
	if k, ok := s.paths[name]; ok {
		if kn, b := k.(*knot); b {
			return kn
		}
	}

	return nil
}

func (s *Story) divert(path string, from Node) Node {
	sp := strings.Split(path, ".")
	kn, st := s.container(from)

	switch len(sp) {
	case 1: // local label || local stitch || story's knot
		if strings.ToLower(path) == "end" {
			return s.end
		}
		// local label
		if kn != nil && st != nil {
			p := kn.name + PathSplit + st.name + PathSplit + path
			if s.paths[p] != nil {
				return s.paths[p]
			}
		}
		// find local stitch
		if kn != nil && kn.stitch(path) != nil {
			return kn.stitch(path)
		}
		// global knot
		if s.knot(path) != nil {
			return s.knot(path)
		}
	case 2: // local stitch.label || knot.stitch
		if kn != nil {
			p := regReplaceDot.ReplaceAllString(path, PathSplit+"$1")
			p = kn.name + PathSplit + p
			if s.paths[p] != nil {
				return s.paths[p]
			}
		}
		if k := s.knot(sp[0]); k != nil {
			return k.stitch(sp[1])
		}
	default: // could be - knot.stitch.label
		p := regReplaceDot.ReplaceAllString(path, PathSplit+"$1")
		// fmt.Println(path)
		return s.paths[p]
	}
	return nil
}

// Context of the story
type Context struct {
	current string
	vars    map[string]interface{}
}

// NewContext which starts from beginning with empty vars
func NewContext() *Context {
	return &Context{
		current: "start",
		vars:    make(map[string]interface{}),
	}
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
type nodes []Node

// NewSection creates the rendered result of the nodes
func (n nodes) section() *Section {
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
	// text = "[start]"
	tags = append(tags, "START")
	return
}

type end struct {
	*base
}

func (e *end) End() (text string, tags []string) {
	// text = "[start]"
	tags = append(tags, "END")
	return
}

// Default story
func Default() *Story {
	parsers := []ParseFunc{readKnot, readStitch, readOption, readGather, readLine}

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

// Parse the input text
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
				if err != errNotMatch {
					return err
				}
			} else {
				break
			}
		}
	}

	return nil
}

func (s *Story) PostParsing() error {
	for _, node := range s.paths {
		if err := node.PostParsing(); err != nil {
			return err
		}
	}
	return nil
}
