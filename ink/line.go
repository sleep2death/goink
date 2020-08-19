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

	gatherReg    = regexp.MustCompile(`^((-\s*)+)([^>].+)`)
	labelReg     = regexp.MustCompile(`^\s*\((.+)\)(.*)`)
	validNameReg = regexp.MustCompile(`^[a-zA-Z_]\w*$`)
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
	if len(i.tags) > 1 {
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

	// TODO: exprc parsing

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
	p := l.parent
	for p != nil {
		if c, ok := p.(*options); ok {
			if c.gather != nil {
				return c.gather, nil
			}
		}

		p = p.Parent()
	}

	return nil, errors.Errorf("current node can not go next: [%s]", l.Path())
}

func (l *line) Render() (text string, tags []string) {
	return l.text, l.tags
}

func (l *line) parseLabel() error {
	if res := labelReg.FindStringSubmatch(l.text); res != nil {
		label := strings.TrimSpace(res[1])
		if len(label) > 0 {
			if valid := validNameReg.FindString(label); valid == "" {
				return errors.Errorf("invalid label name: %s", label)
			}

			if knot, stitch := l.story.container(l); stitch != nil {
				label = stitch.Path() + PathSplit + label
			} else if knot != nil {
				label = knot.Path() + PathSplit + label
			}

			if _, ok := l.story.paths[label]; ok {
				return errors.Errorf("conflict label name: %s", label)
			}
			l.story.paths[label] = l
			l.path = label
		}
		l.text = res[2]
	}

	return nil
}

// readGather create and insert a new gather into story
func readGather(s *Story, input string) error {
	res := gatherReg.FindStringSubmatch(input)
	if res != nil {
		nesting := len(strings.Join(strings.Fields(res[1]), ""))
		i, err := newLine(res[3])
		if err != nil {
			return err
		}

		g := &gather{line: i, nesting: nesting}
		g.story = s

		// parsing label
		if err := i.parseLabel(); err != nil {
			return err
		}

		obj := s.current
		var choices *options
		for obj != nil {
			if c, ok := obj.(*options); ok {
				if t := nesting - c.nesting; t == 0 {
					choices = c
					break
				}
			}

			obj = obj.Parent()
		}

		if choices != nil && choices.gather == nil {
			// if gather didn't have label
			// set the default path
			if g.path == "" {
				g.path = choices.Path() + PathSplit + "g"
				s.paths[g.path] = g
			}

			g.parent = nil // forbid gather from parenting

			choices.gather = g
			s.current = g

			return nil
		}

		return errors.Errorf("cannot find the choices of the gather: %s", input)
	}

	return errNotMatch
}

// gather node of the choices
type gather struct {
	*line
	nesting int
}
