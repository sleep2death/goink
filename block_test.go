package goink

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadLines(t *testing.T) {
	_, err := readInk("./ink/sample.ink")
	assert.Nil(t, err)

	// t.Log(len(s.children))
	// assert.Equal(t, 9, len(lines))
}

func TestNestedChoice(t *testing.T) {
	ls := `
	Hello, world!
	-> Knot_1
	== Knot_1
	  Knot_1 content.
	  + Choice_1
		Choice 1 Content
		++ Choice_1.1
		  Choice 1.1 Content.
		  -> ending
		++ Choice_1.2
		  Choice 1.2 Content.
		  -> ending
	  + Choice_2
		Choice 2 Content.
		+++ Choice_2.1
		  Choice 2.2 Content.
		  -> ending
		++ Choice_2.2 // less than prev choice but bigger than parent
		  Choice 2.2 Content.
		  -> ending
	== ending
	  This is the end.
	  -> END
	`
	lines := strings.Split(ls, "\n")
	res, err := parseLines(lines)
	assert.Nil(t, err)
	t.Log(res.format(""))
	compare(t, ls, res)
}

func parseLines(lines []string) (blk *block, err error) {
	blk = &block{}
	i := 0
	for {
		if i < len(lines) {
			blk, err = blk.parse(strings.TrimSpace(lines[i]))
			if err != nil {
				return nil, err
			}
			i++
		} else {
			break
		}
	}
	blk = blk.root()
	return
}

func compare(t *testing.T, expected string, actual *block) {
	assert.Equal(t, strings.ReplaceAll(expected, "\t", "    "), "\n"+actual.format("    ")+"    ")
}

func TestInlineParse(t *testing.T) {
	inline := &inline{raw: "=== hello ==="}
	inline.parse()
}
