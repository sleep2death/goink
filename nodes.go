package goink

import "github.com/pkg/errors"

// Node is the basic element of the story
type Node interface {
	Story() *Story
	Prev() Node
}

// Next is a node which can procceed next
type Next interface {
	SetNext(prev Node)
	Next() Node
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

// PlainText of the story
type PlainText struct {
	s *Story // story
	p Node   // prev
	n Node   // next

	raw string
}

// Story of the content
func (p *PlainText) Story() *Story {
	return p.s
}

// Next content
func (p *PlainText) Next() Node {
	return p.n
}

// SetNext content
func (p *PlainText) SetNext(prev Node) {
	p.n = prev
}

// Prev content
func (p *PlainText) Prev() Node {
	return p.p
}
