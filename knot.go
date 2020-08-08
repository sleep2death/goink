package goink

import (
	"regexp"
	"strings"

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

		if s.FindKnot(name) != nil {
			return errors.Errorf("conflict knot name: %s", name)
		}

		k := &Knot{story: s, name: name}

		s.knots = append(s.knots, k)
		s.current = k

		k.path = "r.k_" + name
		k.ln = s.ln

		// s.objMap[k.path] = k

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

		var k *Knot
		for obj != nil {
			if _, ok := obj.(*Knot); ok {
				k = obj.(*Knot)
				break
			} else if _, ok := obj.(*Stitch); ok {
				k = obj.(*Stitch).knot
				break
			}

			obj = obj.Parent()
		}

		if k == nil {
			return errors.Errorf("can not find the knot of the stitch: %s", input)
		}

		if k.FindStitch(name) != nil {
			return errors.Errorf("conflict stitch name: %s", name)
		}

		stitch := &Stitch{story: s, name: name, knot: k}
		k.stitches = append(k.stitches, stitch)
		s.current = stitch

		stitch.path = k.Path() + ".s_" + name
		stitch.ln = s.ln
		// s.objMap[stitch.path] = stitch

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

// path of the stitch
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

// FindDivert in the given path
func (s *Story) FindDivert(path string, obj InkObj) InkObj {
	split := strings.Split(path, ".")

	switch len(split) {
	case 1:
		// find local stitch or lable first
		if obj != nil {
			var k *Knot
			obj := obj.Parent()

			for obj != nil {
				if _, ok := obj.(*Knot); ok {
					k = obj.(*Knot)
					break
				} else if _, ok := obj.(*Stitch); ok {
					k = obj.(*Stitch).knot
					break
				}

				obj = obj.Parent()
			}

			if k != nil {
				// fix knot and stitch visit count
				// return k.FindStitch(i.divert).Next()
				if stitch := k.FindStitch(path); stitch != nil {
					return stitch
				}
			}
		}
		return s.FindKnot(path)
	case 2:
		if k := s.FindKnot(split[0]); k != nil {
			return k.FindStitch(split[1])
		}
	case 3:
		//TODO: Find label
	}
	return nil
}

// FindDivertCount in the given path
func (s *Story) FindDivertCount(path string) int {
	split := strings.Split(path, ".")

	switch len(split) {
	case 1:
	case 2:
	case 3:
		//TODO: Find label
	}
	return -1
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
