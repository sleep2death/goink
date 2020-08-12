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

	lableReg = regexp.MustCompile(`^\s*\((.+)\)(.*)`)
)

// NewInline parse and insert a new inline into story
func NewInline(s *Story, input string) error {
	i, err := CreateNewInline(input)

	if err != nil {
		return err
	}

	i.story = s
	i.parent = s.current

	i.path = s.current.Path() + split + "i"
	s.objMap[i.path] = i

	s.current.SetNext(i)
	s.current = i

	return nil
}

// CreateNewInline from the input
func CreateNewInline(input string) (*Inline, error) {
	// Inline
	i := &Inline{raw: input}

	// illegal gather sign
	if input[:1] == "-" && input[:2] != "->" {
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

// Inline node of tye story
type Inline struct {
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
func (i *Inline) Render() string {
	return i.text
}

// Parent of the inline
func (i *Inline) Parent() InkObj {
	return i.parent
}

// Path of the inline
func (i *Inline) Path() string {
	return i.path
}

// SetNext content of the inline
func (i *Inline) SetNext(obj InkObj) {
	i.next = obj
}

// Next content of the inline
func (i *Inline) Next() InkObj {
	// divert
	if i.divert != "" {
		// return i.story.FindDivert(i.divert).Next()
		return i.story.FindDivert(i.divert, i)
	}

	// fallback to next
	if i.next != nil {
		return i.next
	}

	// fallback to gather
	obj := i.parent
	for obj != nil {
		if c, ok := obj.(*Choices); ok {
			if c.gather != nil {
				return c.gather
			}
		}

		obj = obj.Parent()
	}

	return nil
}

// Story of the inline
func (i *Inline) Story() *Story {
	return i.story
}
