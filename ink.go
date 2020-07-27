package goink

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Story of the ink
type Story struct {
	start Node // start line of the story
	end   Node
	ln    int

	knots []*Knot

	current Node //current line of the story
}

// Reset the story
func (s *Story) Reset() {
	s.current = s.start
}

// Next content of the story
func (s *Story) Next() (Node, error) {
	if next, ok := s.current.(Next); ok {
		if n := next.Next(); n != nil {
			s.current = n
			return s.current, nil
		}
	}

	return nil, errors.Errorf("cannot go next: %d", s.current.LN())
}

// Select the choice
func (s *Story) Select(idx int) (Node, error) {
	if choices, ok := s.current.(*Choices); ok {
		return choices.Select(idx)
	}
	return nil, errors.Errorf("cannot select: %d", s.current.LN())
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

var (
	choiceReg = regexp.MustCompile(`((^\++)|(^\*+))\s(.+)`)
	knotReg   = regexp.MustCompile(`(^\={2,})\s(\w+)`)
	stitchReg = regexp.MustCompile(`(^\=)\s(\w+)`)
)

// Parse input string into contents
func Parse(s *Story, input string) error {
	s.ln++
	// trim spaces and skip empty lines
	input = strings.TrimRight(strings.TrimSpace(input), "\r\n")
	if len(input) == 0 {
		return nil
	}

	if s.current == nil {
		s.current = s.start
	}

	next, canNext := s.current.(Next)

	if canNext {
		errors.Errorf("current node cannot continue: %d", s.ln)
	}

	// == knot
	result := knotReg.FindStringSubmatch(input)
	if result != nil {
		k := &Knot{s: s, name: result[2], ln: s.ln}
		s.knots = append(s.knots, k)
		s.current = k

		return nil
	}

	// stitch
	result = stitchReg.FindStringSubmatch(input)
	if result != nil {
		stitch := &Stitch{s: s, name: result[2], ln: s.ln}
		inline := s.current

		for {
			if k, ok := inline.(*Knot); ok {
				k.stitches = append(k.stitches, stitch)
				s.current = stitch

				return nil
			}

			if p, ok := inline.(Prev); ok {
				inline = p.Prev()
			} else {
				return errors.Errorf("cannot find stitch's knot: %d", stitch.ln)
			}
		}
	}

	// * choices
	result = choiceReg.FindStringSubmatch(input)
	if result != nil {
		nesting := len(result[2]) + len(result[3])
		// c := &Inline{s: s, raw: result[4]}
		c := NewInline(result[4])
		c.s = s
		c.ln = s.ln
		choices := findChoices(s, nesting)
		if choices == nil {
			choices = &Choices{s: s, p: s.current, nesting: nesting, ln: s.ln}
			next.SetNext(choices)
		}

		// add plain text of the choice into choices,
		// and make it the current node
		c.p = choices
		choices.selections = append(choices.selections, c)

		s.current = c
		return nil
	}

	// plain text
	il := NewInline(input)
	il.s = s
	il.ln = s.ln
	il.p = s.current

	// if inline's divert is not empty, set its knot
	// for local stitch finding
	if len(il.divert) > 0 {
		inline := s.current

		for {
			if k, ok := inline.(*Knot); ok {
				il.k = k
				break
			}

			if p, ok := inline.(Prev); ok {
				inline = p.Prev()
			} else {
				break
			}
		}
	}

	next.SetNext(il)
	s.current = il
	return nil
}

func findChoices(s *Story, nesting int) *Choices {
	inline := s.current

	for {
		if choices, ok := inline.(*Choices); ok {
			if choices.nesting < nesting {
				return nil
			} else if choices.nesting == nesting {
				s.current = choices
				return choices
			}
		}

		if p, ok := inline.(Prev); ok {
			inline = p.Prev()
		} else {
			return nil
		}
	}
}

// NewStory of the Ink
func NewStory() *Story {
	return &Story{}
}
