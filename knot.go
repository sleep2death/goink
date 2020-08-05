package goink

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Knot is a container of story's content
type Knot struct {
	story *Story
	name  string

	stitches []*Stitch
	next     InkObj
}

var knotReg = regexp.MustCompile(`(^\={2,})(\s+)(\w+)`)
var stitchReg = regexp.MustCompile(`(^\=)(\s+)(\w+)`)

// NewKnot parse and insert a new knot into story
func NewKnot(s *Story, input string) error {
	// == knot
	result := knotReg.FindStringSubmatch(input)
	if result != nil {
		name := result[3]

		if s.FindKnot(name) != nil {
			return errors.Errorf("conflict knot name: %s", name)
		}

		k := &Knot{story: s, name: name}
		s.knots = append(s.knots, k)
		s.current = k

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

		obj := s.current
		for obj != nil {
			if k, ok := obj.(*Knot); ok {
				if k.FindStitch(name) != nil {
					return errors.Errorf("conflict stitch name: %s", name)
				}
				stitch := &Stitch{story: s, name: name}
				k.stitches = append(k.stitches, stitch)
				s.current = stitch
				return nil
			}

			obj = obj.Parent()
		}

		return errors.Errorf("can not find the knot of this stitch: %s", input)
	}

	return ErrNotMatch
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
	name  string

	next InkObj
}

// Name of the stitch
func (s *Stitch) Name() string {
	return s.name
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

// FindDivert of the given obj
func (s *Story) FindDivert(obj InkObj, path string) InkObj {
	split := strings.Split(path, ".")
	switch len(split) {
	case 1:
		// local stitch first
		for obj != nil {
			if s, ok := obj.(*Stitch); ok {
				if s.name == path {
					return s.Next()
				}
			}
			obj = obj.Parent()
		}
		return s.FindKnot(path)
	case 2:
		if k := s.FindKnot(split[0]); k != nil {
			if s := k.FindStitch(split[1]); s != nil {
				return s
			}
		}
	case 3:
		//TODO: Find label
	}
	return nil
}

// FindKnot of the story by name
func (s *Story) FindKnot(name string) *Knot {
	for _, k := range s.knots {
		if k.name == name {
			return k
		}
	}
	return nil
}

// FindStitch of the knot by name
func (k *Knot) FindStitch(name string) *Stitch {
	for _, s := range k.stitches {
		if s.name == name {
			return s
		}
	}
	return nil
}
