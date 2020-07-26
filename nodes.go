package goink

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Node is the basic element of the story
type Node interface {
	Story() *Story
}

// Next is a node which can procceed next
type Next interface {
	SetNext(prev Node)
	Next() Node
}

// Prev is a node which can procceed prev
type Prev interface {
	// don't need setter
	Prev() Node
}

// Choices of the story
type Choices struct {
	s *Story // story
	p Node   // prev

	nesting    int
	selections []Node
}

// Story of the content
func (c *Choices) Story() *Story {
	return c.s
}

// Prev content
func (c *Choices) Prev() Node {
	return c.p
}

// Select  content
func (c *Choices) Select(idx int) (Node, error) {
	idx = idx - 1
	if idx < 0 || idx >= len(c.selections) {
		return nil, errors.Errorf("Invalid selection index: %d", idx)
	}

	c.s.current = c.selections[idx]
	return c.s.current, nil
}

// Inline of the story
type Inline struct {
	s *Story // story
	p Node   // prev
	n Node   // next

	comment string
	tags    []string
	divert  string
	text    string

	raw string
}

// Story of the content
func (i *Inline) Story() *Story {
	return i.s
}

// Next content
func (i *Inline) Next() Node {
	return i.n
}

// SetNext content
func (i *Inline) SetNext(next Node) {
	i.n = next
}

// Prev content
func (i *Inline) Prev() Node {
	return i.p
}

var (
	commentReg = regexp.MustCompile(`(^.*)(\/\/)(.+)$`)
	tagReg     = regexp.MustCompile(`(^.*)(\#)(.+)$`)
	divertReg  = regexp.MustCompile(`(^.+)(\-\>)(.+)$`)
)

// NewInline from input
func NewInline(input string) *Inline {
	i := &Inline{raw: input}

	// comment | spaces trimed
	if res := commentReg.FindStringSubmatch(input); res != nil {
		input = res[1]
		i.comment = strings.TrimSpace(res[3])
	}

	// tags | spaces trimmed
	res := tagReg.FindStringSubmatch(input)
	for res != nil {
		input = res[1]
		i.tags = append(i.tags, strings.TrimSpace(res[3]))
		res = tagReg.FindStringSubmatch(input)
	}

	// divert | spaces trimmed
	if res := divertReg.FindStringSubmatch(input); res != nil {
		input = res[1]
		i.divert = strings.TrimSpace(res[3])
	}

	// text | spaces not trimmed
	i.text = input

	return i
}
