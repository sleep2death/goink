package goink

import (
	"regexp"

	"github.com/pkg/errors"
)

var knotReg = regexp.MustCompile(`(^\={2,})(\s+)(\w+)`)
var stitchReg = regexp.MustCompile(`(^\=)(\s+)(\w+)`)

// readKnot parse and insert a new knot into story
func readKnot(s *Story, input string) error {
	// == knot
	result := knotReg.FindStringSubmatch(input)
	if result != nil {
		name := result[3]

		k := &knot{story: s, name: name}
		s.knots = append(s.knots, k)
		s.current = k

		k.path = name

		if s.objMap[k.path] != nil {
			return errors.Errorf("conflict knot name: %s", name)
		}
		s.objMap[k.path] = k

		return nil
	}

	return ErrNotMatch
}

// readStitch parse and insert a new knot into story
func readStitch(s *Story, input string) error {
	// = stitch
	result := stitchReg.FindStringSubmatch(input)
	if result != nil {
		name := result[3]

		k, _ := s.findContainer(s.current)
		if k == nil {
			return errors.Errorf("can not find the knot of the stitch: %s", input)
		}

		if k.findStitch(name) != nil {
			return errors.Errorf("conflict stitch name: %s", name)
		}

		stitch := &Stitch{story: s, name: name, k: k}
		k.stitches = append(k.stitches, stitch)
		s.current = stitch

		stitch.path = k.Path() + SPLIT + name

		// do not need check again
		/* if s.objMap[stitch.path] != nil {
			return errors.Errorf("conflict stitch name: %s", name)
		} */
		s.objMap[stitch.path] = stitch

		return nil
	}

	return ErrNotMatch
}

// knot is a container of story's content
type knot struct {
	story *Story
	name  string

	path     string
	stitches []*Stitch
	next     InkObj
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
func (k *knot) Parent() InkObj {
	return nil
}

// SetNext of the knot
func (k *knot) SetNext(obj InkObj) {
	k.next = obj
}

// Next of the knot
func (k *knot) Next() InkObj {
	return k.next
}

// Stitch is a sub container of a knot
type Stitch struct {
	story *Story
	k     *knot
	name  string

	path string
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

// findStitch of the knot by name
func (k *knot) findStitch(name string) *Stitch {
	if s, ok := k.story.objMap[k.name+SPLIT+name]; ok {
		if stitch, b := s.(*Stitch); b {
			return stitch
		}
	}

	return nil
}
