package goink

import (
	"regexp"

	"github.com/pkg/errors"
)

var knotReg = regexp.MustCompile(`(^\={2,})(\s+)(\w+)`)
var stitchReg = regexp.MustCompile(`(^\=)(\s+)(\w+)`)

// NewKnot parse and insert a new knot into story
func NewKnot(s *Story, input string) error {
	// == knot
	result := knotReg.FindStringSubmatch(input)
	if result != nil {
		name := result[3]

		k := &Knot{story: s, name: name}
		s.knots = append(s.knots, k)
		s.current = k

		k.path = name
		k.ln = s.ln

		if s.objMap[k.path] != nil {
			return errors.Errorf("conflict knot name: %s", name)
		}
		s.objMap[k.path] = k

		return nil
	}

	return ErrNotMatch
}

// NewStitch parse and insert a new knot into story
func NewStitch(s *Story, input string) error {
	// = stitch
	result := stitchReg.FindStringSubmatch(input)
	if result != nil {
		name := result[3]

		k, _ := s.FindContainer(s.current)
		if k == nil {
			return errors.Errorf("can not find the knot of the stitch: %s", input)
		}

		if k.FindStitch(name) != nil {
			return errors.Errorf("conflict stitch name: %s", name)
		}

		stitch := &Stitch{story: s, name: name, knot: k}
		k.stitches = append(k.stitches, stitch)
		s.current = stitch

		stitch.path = k.Path() + "." + name
		stitch.ln = s.ln

		if s.objMap[stitch.path] != nil {
			return errors.Errorf("conflict stitch name: %s", name)
		}
		s.objMap[stitch.path] = stitch

		return nil
	}

	return ErrNotMatch
}

// Knot is a container of story's content
type Knot struct {
	story *Story
	name  string

	path string
	ln   int

	stitches []*Stitch
	next     InkObj
}

// Path of the knot
func (k *Knot) Path() string {
	return k.path
}

// Name of the knot
func (k *Knot) Name() string {
	return k.name
}

// Story of the knot
func (k *Knot) Story() *Story {
	return k.story
}

// Parent of the knot should always be nil
func (k *Knot) Parent() InkObj {
	return nil
}

// SetNext of the knot
func (k *Knot) SetNext(obj InkObj) {
	k.next = obj
}

// Next of the knot
func (k *Knot) Next() InkObj {
	return k.next
}

// Stitch is a sub container of a knot
type Stitch struct {
	story *Story
	knot  *Knot
	name  string

	path string
	ln   int

	next InkObj
}

// Name of the stitch
func (s *Stitch) Name() string {
	return s.name
}

// Path of the stitch
func (s *Stitch) Path() string {
	return s.path
}

// Story of the stitch
func (s *Stitch) Story() *Story {
	return s.story
}

// Parent of the stitch should always be nil
func (s *Stitch) Parent() InkObj {
	return nil
}

// SetNext of the stitch
func (s *Stitch) SetNext(obj InkObj) {
	s.next = obj
}

// Next of the stitch
func (s *Stitch) Next() InkObj {
	return s.next
}

// FindStitch of the knot by name
func (k *Knot) FindStitch(name string) *Stitch {
	if s, ok := k.story.objMap[k.name+"."+name]; ok {
		if stitch, b := s.(*Stitch); b {
			return stitch
		} else {
			panic(errors.Errorf("type error with the name: %s", name))
		}
	}

	return nil
}
