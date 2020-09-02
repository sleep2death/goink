package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScannerInit(t *testing.T) {
	input := `
你好, ABC
World
	`
	s := &scanner{}
	s.init([]byte(input), nil)
	pos, lit := s.scan()

	t.Log(pos, lit)
	assert.Equal(t, 0, len(s.lines))
}
