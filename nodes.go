package goink

import (
	"github.com/pkg/errors"
)

// Node is the basic element of the story
type Node interface {
	Story() *Story
	LN() int // LineNumber
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
	s  *Story // story
	p  Node   // prev
	ln int    // line number

	nesting    int
	selections []Node
	gather     *Gather
}

// Story of the content
func (c *Choices) Story() *Story {
	return c.s
}

// LN - line number of the content
func (c *Choices) LN() int {
	return c.ln
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

// Knot of the story
type Knot struct {
	s  *Story
	n  Node
	ln int

	name     string // name of the knot
	stitches []*Stitch
}

// Story of the knot
func (k *Knot) Story() *Story {
	return k.s
}

// LN - line number of the content
func (k *Knot) LN() int {
	return k.ln
}

// Next content
func (k *Knot) Next() Node {
	return k.n
}

// SetNext content
func (k *Knot) SetNext(next Node) {
	k.n = next
}

// FindStitch of the knot
func (k *Knot) FindStitch(name string) *Stitch {
	for _, s := range k.stitches {
		if s.name == name {
			return s
		}
	}
	return nil
}

// Stitch of the knot
type Stitch struct {
	s  *Story
	n  Node
	k  *Knot
	ln int

	name string // name of the knot
}

// Story of the stitch
func (s *Stitch) Story() *Story {
	return s.s
}

// LN - line number of the content
func (s *Stitch) LN() int {
	return s.ln
}

// Next content
func (s *Stitch) Next() Node {
	return s.n
}

// SetNext content
func (s *Stitch) SetNext(next Node) {
	s.n = next
}

// Gather of the choices
type Gather struct {
	*Inline
	nesting int
}
