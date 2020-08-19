package ink

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var (
	commentReg = regexp.MustCompile(`(^.*)(\/\/)(.+)$`)
	tagReg     = regexp.MustCompile(`(^.*)(\#)(.+)$`)
	divertReg  = regexp.MustCompile(`(^.*)(\-\>)(.+)$`)

	glueStartReg = regexp.MustCompile(`^\<\>(.+)`)
	glueEndReg   = regexp.MustCompile(`(.+)\<\>$`)

	// illegalGatherReg = regexp.MustCompile(`\-\-\>`)
)

// readLine parse and insert a new inline into story
func readLine(s *Story, input string) error {
	l, err := newLine(input)

	if err != nil {
		return err
	}

	l.story = s
	l.parent = s.current

	l.path = s.current.Path() + PathSplit + "i"
	s.paths[l.path] = l

	if n := s.next(); n != nil {
		n.SetNext(l)
		s.current = l

		return nil
	}
	return errors.Errorf("current node can not set next: %s", s.current.Path())
}

// newLine from the input
func newLine(input string) (*line, error) {
	// Inline
	i := &line{base: &base{}, raw: input}

	// illegal gather sign
	/* if res := illegalGatherReg.FindStringSubmatch(input); res != nil {
		return nil, errors.Errorf("illegal gather character: %s", input)
	} */

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
		i.divert = strings.ToLower(strings.TrimSpace(res[3]))
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
	*base

	next Node
	raw  string

	comment string
	tags    []string
	divert  string

	glueStart bool
	glueEnd   bool

	text string
}

// SetNext content of the inline
func (l *line) SetNext(obj Node) {
	l.next = obj
}

// Next content of the inline
func (l *line) Next() (Node, error) {
	// divert
	if l.divert != "" {
		if target := l.story.divert(l.divert, l); target != nil {
			return target, nil
		}

		return nil, errors.Errorf("can not find the divert: %s", l.divert)
	}

	// fallback to next
	if l.next != nil {
		return l.next, nil
	}

	// fallback to gather
	/* p := l.parent
	for p != nil {
		if c, ok := p.(*options); ok {
			if c.gather != nil {
				return c.gather
			}
		}

		p = p.Parent()
	} */

	return nil, errors.Errorf("current node can not go next: [%s]", l.Path())
}

func (l *line) Render() (text string, tags []string) {
	return l.text, l.tags
}
