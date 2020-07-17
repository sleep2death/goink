package goink

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"
)

type lineHeader int

const (
	lhString lineHeader = iota
	lhKnot
	lhChoicePlus
	lhChoiceStar
	lhDevert
)

// ReadLines from ink file.
func readLines(path string) (lines []line, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	reader := bufio.NewReader(file)
	buffer := bytes.NewBuffer(make([]byte, 1024))

	for {
		part, prefix, err := reader.ReadLine()
		if err != nil {
			break
		}

		buffer.Write(part)

		if !prefix {
			// skip empty line
			if len(buffer.String()) > 0 {
				line, err := readLine(strings.TrimSpace(buffer.String()))
				if err != nil {
					return lines, err
				}
				lines = append(lines, line)
			}
			buffer.Reset()
		}
	}

	if err == io.EOF {
		err = nil
	}

	return
}

type line struct {
	header    lineHeader
	subString string
}

func readLine(str string) (l line, err error) {
	r := regexp.MustCompile(`^(\*\s)|^(\+\s)|^(={2,}\s)|^(\-\>\s)`)
	res := r.FindStringSubmatch(str)

	if len(res) > 0 {
		for i := 1; i < len(res); i++ {
			if len(res[i]) > 0 {
				switch i {
				case 1:
					return line{header: lhChoiceStar}, nil
				case 2:
					return line{header: lhChoicePlus}, nil
				case 3:
					return line{header: lhKnot}, nil
				case 4:
					return line{header: lhDevert}, nil
				}
			}
		}
	}
	return line{header: lhString, subString: str}, nil
}
