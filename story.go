package goink

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type header int

const (
	empty header = iota
	text
	knot
	choice
)

func (lh header) String() string {
	return [...]string{"empty", "text", "knot", "choice"}[lh]
}

// ReadLines from ink file.
func readInk(path string) (story *Story, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	reader := bufio.NewReader(file)
	buffer := bytes.NewBuffer(make([]byte, 1024))

	story = &Story{}
	for {
		part, prefix, err := reader.ReadLine()
		if err != nil {
			break
		}

		buffer.Write(part)

		if !prefix {
			err = story.parse(strings.TrimSpace(buffer.String()))

			if err != nil {
				story.status = psError
				return story, err
			}

			buffer.Reset()
		}
	}

	if err == io.EOF {
		err = nil
	}

	story.status = psEnd
	return
}

var headerReg = regexp.MustCompile(`^(\+\s)|^(={2,}\s)`)

// ParseStatus of the inkobj
type parseStatus int

const (
	psError parseStatus = iota - 1
	psStart
	psEnd
)

// InkObj is the basic element of ink story
type InkObj interface {
	parent() InkObj
	content() []InkObj

	parse(line string) error
	parseStatus() parseStatus
}

// Story of the ink file
type Story struct {
	c       []InkObj
	status  parseStatus
	lineNum int
}

func (s *Story) parent() InkObj {
	return nil
}

func (s *Story) content() []InkObj {
	return s.c
}

func (s *Story) parseStatus() parseStatus {
	return s.status
}

// parse story from input lines
func (s *Story) parse(line string) error {
	s.lineNum++
	h := empty
	if len(line) > 0 {
		str := headerReg.FindStringSubmatch(line)
		if len(str) == 0 {
			h = text
			t := &Line{p: s, lineNum: s.lineNum}
			s.c = append(s.c, t)
		} else if len(str[1]) > 0 {
			h = choice
		} else if len(str[2]) > 0 {
			h = knot
		}
	}
	fmt.Printf("[%d] %s\n", s.lineNum, h)
	return nil
}

// Line parsing struct
type Line struct {
	p       InkObj
	c       []InkObj
	status  parseStatus
	lineNum int
}

func (t *Line) parent() InkObj {
	return t.p
}

func (t *Line) content() []InkObj {
	return t.c
}

func (t *Line) parse(line string) error {
	return nil
}

func (t *Line) parseStatus() parseStatus {
	return t.status
}
