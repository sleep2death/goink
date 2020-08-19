package ink

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var (
	knotReg   = regexp.MustCompile(`(^\={2,})(\s+)(\w+)`)
	stitchReg = regexp.MustCompile(`(^\=)(\s+)(\w+)`)
)

// readKnot parse and insert a new knot into story
func readKnot(s *Story, input string) error {
	// == knot
	result := knotReg.FindStringSubmatch(input)
	if result != nil {
		name := strings.ToLower(result[3])

		k := &knot{story: s, name: name}
		s.knots = append(s.knots, k)
		s.current = k

		k.path = name

		if s.paths[k.path] != nil {
			return errors.Errorf("conflict knot name: %s", name)
		}
		s.paths[k.path] = k

		return nil
	}

	return errNotMatch
}

// knot is a container of story's content
type knot struct {
	story *Story
	path  string

	stitches []*stitch
	name     string
	next     Node
}

// Path of the knot
func (k *knot) Path() string {
	return k.path
}

// Name of the knot
func (k *knot) Name() string {
	return k.name
}

// Story of the knot
func (k *knot) Story() *Story {
	return k.story
}

// Parent of the knot should always be nil
func (k *knot) Parent() Node {
	return nil
}

// SetNext of the knot
func (k *knot) SetNext(obj Node) {
	k.next = obj
}

// Next of the knot
func (k *knot) Next() (Node, error) {
	return k.next, nil
}

// Render the content of knot... should be both empty
func (k *knot) Render() (output string, tags []string) {
	return "", nil
}

// find stitch of the knot by name
func (k *knot) stitch(name string) *stitch {
	if s, ok := k.story.paths[k.name+PathSplit+name]; ok {
		if stitch, b := s.(*stitch); b {
			return stitch
		}
	}

	return nil
}

// readStitch parse and insert a new knot into story
func readStitch(s *Story, input string) error {
	// = stitch
	result := stitchReg.FindStringSubmatch(input)
	if result != nil {
		name := strings.ToLower(result[3])

		k, _ := s.container(s.current)
		if k == nil {
			return errors.Errorf("can not find the knot of the stitch: %s", input)
		}

		if k.stitch(name) != nil {
			return errors.Errorf("conflict stitch name: %s", name)
		}

		stitch := &stitch{story: s, name: name, knot: k}
		k.stitches = append(k.stitches, stitch)
		s.current = stitch

		stitch.path = k.Path() + PathSplit + name
		s.paths[stitch.path] = stitch

		return nil
	}

	return errNotMatch
}

// stitch is a sub container of a knot
type stitch struct {
	story *Story
	knot  *knot
	name  string

	path string
	next Node
}

// Name of the stitch
func (s *stitch) Name() string {
	return s.name
}

// Path of the stitch
func (s *stitch) Path() string {
	return s.path
}

// Story of the stitch
func (s *stitch) Story() *Story {
	return s.story
}

// Parent of the stitch should always be nil
func (s *stitch) Parent() Node {
	return nil
}

// SetNext of the stitch
func (s *stitch) SetNext(obj Node) {
	s.next = obj
}

// Next of the stitch
func (s *stitch) Next() (Node, error) {
	return s.next, nil
}

// Render the content of stitch... should be both empty
func (s *stitch) Render() (output string, tags []string) {
	return "", nil
}
