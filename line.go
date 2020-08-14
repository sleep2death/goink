package goink

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// InkObj is the basic element of the story
type InkObj interface {
	Story() *Story
	Parent() InkObj

	Next() InkObj
	SetNext(obj InkObj)

	Path() string
}

var (
	commentReg = regexp.MustCompile(`(^.*)(\/\/)(.+)$`)
	tagReg     = regexp.MustCompile(`(^.*)(\#)(.+)$`)
	divertReg  = regexp.MustCompile(`(^.*)(\-\>)(.+)$`)

	glueStartReg = regexp.MustCompile(`^\<\>(.+)`)
	glueEndReg   = regexp.MustCompile(`(.+)\<\>$`)

	lableReg         = regexp.MustCompile(`^\s*\((.+)\)(.*)`)
	illegalGatherReg = regexp.MustCompile(`\-\-\>`)
)

// readLine parse and insert a new inline into story
func readLine(s *Story, input string) error {
	l, err := newLine(input)

	if err != nil {
		return err
	}

	l.story = s
	l.parent = s.current

	l.path = s.current.Path() + SPLIT + "i"
	s.objMap[l.path] = l

	s.current.SetNext(l)
	s.current = l

	return nil
}

// newLine from the input
func newLine(input string) (*line, error) {
	// Inline
	i := &line{raw: input}

	// illegal gather sign
	if res := illegalGatherReg.FindStringSubmatch(input); res != nil {
		return nil, errors.Errorf("illegal gather character: %s", input)
	}

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
	// reverse tag list
	if len(i.tags) > 0 {
		for j, k := 0, len(i.tags)-1; j < k; j, k = j+1, k-1 {
			i.tags[j], i.tags[k] = i.tags[k], i.tags[j]
		}
	}

	// divert | spaces trimmed
	if res := divertReg.FindStringSubmatch(input); res != nil {
		input = res[1]
		i.divert = strings.TrimSpace(res[3])
	}

	// glue
	if res := glueStartReg.FindStringSubmatch(input); res != nil {
		i.glueStart = true
		input = res[1]
	}

	if res := glueEndReg.FindStringSubmatch(input); res != nil {
		i.glueEnd = true
		input = res[1]
	}

	// text | spaces not trimmed
	i.text = input
	return i, nil
}

// line node of the story
type line struct {
	story  *Story
	parent InkObj
	next   InkObj
	path   string

	raw string

	comment string
	tags    []string
	divert  string

	glueStart bool
	glueEnd   bool

	text string
}

// Render the inline's content into string
func (l *line) Render() string {
	return l.text
}

// Parent of the line
func (l *line) Parent() InkObj {
	return l.parent
}

// Path of the inline
func (l *line) Path() string {
	return l.path
}

// SetNext content of the inline
func (l *line) SetNext(obj InkObj) {
	l.next = obj
}

// Next content of the inline
func (l *line) Next() InkObj {
	// divert
	if l.divert != "" {
		// return i.story.FindDivert(i.divert).Next()
		if target := l.story.findDivert(l.divert, l); target != nil {
			return target
		}

		panic(errors.Errorf("can not find the divert: %s", l.divert))
	}

	// fallback to next
	if l.next != nil {
		return l.next
	}

	// fallback to gather
	obj := l.parent
	for obj != nil {
		if c, ok := obj.(*options); ok {
			if c.gather != nil {
				return c.gather
			}
		}

		obj = obj.Parent()
	}

	return nil
}

// Story of the inline
func (l *line) Story() *Story {
	return l.story
}
