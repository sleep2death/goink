package goink

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// ReadInk file from the given path
func ReadInk(path string) (b Block, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	reader := bufio.NewReader(file)
	buffer := bytes.NewBuffer(make([]byte, 1024))

	b = &Story{}
	for {
		part, prefix, err := reader.ReadLine()
		if err != nil {
			break
		}

		buffer.Write(part)

		if !prefix {
			b, err = b.parse(strings.TrimRight(strings.TrimSpace(buffer.String()), "\r\n"))
			if err != nil {
				return nil, err
			}

			buffer.Reset()
		}
	}

	if err == io.EOF {
		err = nil
	}
	return
}

// Block of the ink
type Block interface {
	Parent() Block
	Content() []Block

	parse(src string) (Block, error)
}

// Story of the ink
type Story struct {
	content []Block
}

// Parent of the story should always be nil
func (s *Story) Parent() Block {
	return nil
}

// Content of the story
func (s *Story) Content() []Block {
	return s.content
}

func (s *Story) parse(raw string) (Block, error) {
	// -- container header --
	// == KNOT ==
	if knot := newKnot(raw); knot != nil {
		if s.findKnot(knot.name) != nil {
			return nil, errors.Errorf("knot name conflict: [%s]", knot.name)
		}
		knot.story = s
		s.content = append(s.content, knot)
		return knot, nil
	}
	// +* CHOICE *+
	if choice := newChoice(raw); choice != nil {
		// diff from original ink, force check nesting level of choices
		// for better reading and writing
		if choice.nesting > 1 {
			return nil, errors.Errorf("nesting of the story choice should always be 1: %s", raw)
		}

		choice.parent = s
		s.content = append(s.content, choice)
		return choice, nil
	}

	//  INLINE
	inline := newInline(raw)
	inline.parent = s
	s.content = append(s.content, inline)
	return s, nil
}

func (s *Story) findKnot(name string) *Knot {
	for _, blk := range s.content {
		k, ok := blk.(*Knot)
		if ok && k.name == name {
			return k
		}
	}

	return nil
}
