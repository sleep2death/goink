package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScannerInit(t *testing.T) {
	input := `
你/好, ABC
World // this is a comment
多行注释 /* multiple lines
comment */
`
	s := &scanner{}
	s.init([]byte(input), nil)

	var str string
	var spos pos

loop:
	for {
		pos, tok, lit := s.scan()

		switch tok {
		case CHAR:
			if str == "" {
				spos = pos
			}
			str += lit
		case STRING:
			if str == "" {
				spos = pos
			}
			str += lit
		case COMMENT:
			t.Log(spos, str)
			str = ""
		case EOF:
			break loop
		}
	}

	assert.Equal(t, 4, len(s.lines))
}
