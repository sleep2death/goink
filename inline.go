package goink

import (
	"regexp"
	"strings"
)

// Inline of the story
type Inline struct {
	s  *Story // story
	k  *Knot  // knot
	p  Node   // prev
	n  Node   // next
	ln int    // line number

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

// LN - line number of the content
func (i *Inline) LN() int {
	return i.ln
}

// Next content
func (i *Inline) Next() Node {
	// divert
	if i.divert != "" {
		// try to find local stitch first
		if i.k != nil {
			if d := i.k.FindStitch(i.divert); d != nil {
				return d.Next()
			}
		}
		// find knot
		// TODO: sub stitch in knot
		if d := i.s.FindKnot(i.divert); d != nil {
			return d.Next()
		}
		// panic(errors.Errorf("cannot find the knot: %s", i.divert))
		return nil
	}

	if i.n != nil {
		return i.n
	}

	// if divert is nil, try to find gather
	p := i.Prev()
	for {
		if c, ok := p.(*Choices); ok && c.gather != nil && c.gather.LN() > i.ln {
			return c.gather
		}
		if _, ok := p.(Prev); ok {
			p = p.(Prev).Prev()
		} else {
			return nil
		}
	}
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
	divertReg  = regexp.MustCompile(`(^.*)(\-\>)(.+)$`)
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
	if len(i.tags) > 0 {
		// reverse tag list
		for j, k := 0, len(i.tags)-1; j < k; j, k = j+1, k-1 {
			i.tags[j], i.tags[k] = i.tags[k], i.tags[j]
		}
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
