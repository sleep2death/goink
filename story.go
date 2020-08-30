package goink

import (
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

var (
	// PathSplit of the node's path
	PathSplit   string = "__"
	errNotMatch error  = errors.New("RegExp Not Match")
)

// Node is the basic element of a story
type Node interface {
	Story() *Story
	Parent() Node
	PostParsing() error

	Path() string
	LN() int
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

// ErrInk for transporting the error info
type ErrInk struct {
	LN      int    `json:"ln" binding:"required"`
	Message string `json:"msg" binding:"required"`
}

// Wrap errors
func (e ErrInk) Wrap(err error) []error {
	msg := ErrInk{LN: -1, Message: err.Error()}
	return []error{&msg}
}

func wrapError(err error, ln int) *ErrInk {
	return &ErrInk{LN: ln, Message: err.Error()}
}

func (e *ErrInk) Error() string {
	return e.Message + " ln: " + strconv.Itoa(e.LN)
}

// embeding struct which implements Node
type base struct {
	story  *Story
	parent Node
	path   string
	ln     int
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

// LN - line number of the node
func (b *base) LN() int {
	return b.ln
}

// do some post parsing check
func (b *base) PostParsing() error {
	return nil
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

	// current parsing line
	ln int
}

// Resume the story
func (s *Story) Resume(ctx *Context) (sec *Section, err *ErrInk) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if err := s.load(ctx); err != nil {
		return nil, wrapError(err, -1)
	}

	if sec, err = s.resume(); err != nil {
		return nil, err
	}

	// update ctx
	*ctx = s.save()
	return
}

// Pick the option
func (s *Story) Pick(ctx *Context, idx int) (sec *Section, err *ErrInk) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if err := s.load(ctx); err != nil {
		return nil, wrapError(err, -1)
	}

	if sec, err = s.pick(idx); err != nil {
		return nil, err
	}
	// update ctx
	*ctx = s.save()
	return
}

// SetID of the story
func (s *Story) SetID(id string) {
	s.id = id
}

// resume the story
func (s *Story) resume() (sec *Section, err *ErrInk) {
	var ns nodes
loop:
	for {
		if s.current == nil {
			return nil, wrapError(errors.New("current node is nil"), -1)
		}

		ns = append(ns, s.current)
		if err := s.visit(s.current.Path()); err != nil {
			return nil, wrapError(err, s.current.LN())
		}

		switch node := s.current.(type) {
		case End:
			break loop
		case Choices:
			break loop
		case CanNext:
			n, err := node.Next()
			if err != nil {
				return nil, wrapError(err, s.current.LN())
			}
			s.current = n
		default:
			return nil, wrapError(errors.New("current line is not recgonized"), -1)
		}
	}

	return ns.section(), nil
}

// pick one of the current choices' option,
// and resume
func (s *Story) pick(idx int) (sec *Section, erri *ErrInk) {
	if c := s.choices(); c != nil {
		if opt, err := c.Pick(idx); err != nil {
			return nil, wrapError(err, c.(Node).LN())
		} else {
			s.current = opt
			return s.resume()
		}
	}
	return nil, wrapError(errors.New("current line is not an option"), s.current.LN())
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
	n, ok := s.paths[ctx.Current]
	if !ok {
		return errors.Errorf("current path [%s] is not existed", ctx.Current)
	}

	s.current = n
	s.vars = ctx.Vars
	return nil
}

func (s *Story) save() Context {
	return Context{Current: s.current.Path(), Vars: copy(s.vars), LN: s.current.LN()}
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
	Current string                 `json:"current" binding:"required"`
	LN      int                    `json:"ln" binding:"required"`
	Vars    map[string]interface{} `json:"vars"`
}

// NewContext which starts from beginning with empty vars
func NewContext() *Context {
	return &Context{
		Current: "start",
		Vars:    make(map[string]interface{}),
		LN:      0,
	}
}

// Section is rendered result of current story
type Section struct {
	Text string   `json:"text"`
	Tags []string `json:"tags"`

	Opts     []string   `json:"opts"`
	OptsTags [][]string `json:"optsTags"`

	End bool `json:"end" binding:"required"`
}

func (s *Section) add(text string, tags []string) {
	if text != "" {
		var tail, header bool
		// glue header
		if res := glueStartReg.FindStringSubmatch(text); res != nil {
			text = res[1]
			header = true
		}

		if s.Text != "" {
			// glue tail
			if res := glueEndReg.FindStringSubmatch(s.Text); res != nil {
				s.Text = res[1]
				tail = true
			}

			if tail || header {
				s.Text = s.Text + text
			} else {
				s.Text = s.Text + "\n" + text
			}
		} else {
			s.Text = text
		}
	}

	if len(tags) > 0 {
		s.Tags = append(s.Tags, tags...)
	}
}

// Nodes list
type nodes []Node

// NewSection creates the rendered result of the nodes
func (n nodes) section() *Section {
	sec := &Section{}
	for _, node := range n {
		switch node := node.(type) {
		case End:
			sec.End = true
			sec.add(node.End())
		case Choices:
			opts, optsTags := node.List()

			sec.Opts = opts
			sec.OptsTags = optsTags
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
		return nil, errors.New("current line can not go next")
	}

	return s.next, nil
}

func (s *start) Render() (text string, tags []string) {
	text = ""
	tags = append(tags, "START")
	return
}

type end struct {
	*base
}

func (e *end) End() (text string, tags []string) {
	text = ""
	tags = append(tags, "END")
	return
}

// Default story
func Default() *Story {
	parsers := []ParseFunc{readVariable, readKnot, readStitch, readOption, readGather, readLine}

	s := &start{base: &base{path: "start"}}
	e := &end{base: &base{path: "end"}}
	s.SetNext(e)

	story := &Story{start: s, end: e, parsers: parsers}

	s.story = story
	e.story = story

	story.paths = make(map[string]Node)
	story.vars = make(map[string]interface{})
	story.ln = 0

	story.paths["start"] = s
	story.paths["end"] = e

	story.current = s
	return story
}

// ParseFunc of the story
type ParseFunc func(s *Story, input string, ln int) error

// Parse the input text
func (s *Story) Parse(input string) *ErrInk {
	contents := strings.Split(input, "\n")
	for _, line := range contents {
		s.ln++

		// trim spaces and skip empty lines
		l := strings.TrimRight(strings.TrimSpace(line), "\r\n")
		if len(l) == 0 {
			continue
		}

		// passing raw input into parsers
		for _, parser := range s.parsers {
			if err := parser(s, l, s.ln); err != nil {
				if err != errNotMatch {
					return wrapError(err, s.ln)
				}
			} else {
				break
			}
		}
	}

	return nil
}

// PostParsing when all input parsing has done
func (s *Story) PostParsing() (errs []*ErrInk) {
	for _, node := range s.paths {
		if e := node.PostParsing(); e != nil {
			errs = append(errs, wrapError(e, node.LN()))
		}
	}
	return
}
