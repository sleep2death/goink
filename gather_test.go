package goink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGather(t *testing.T) {
	input := `
	Hello World,
	- --> Invalid New Gather
	`
	_, err := parse(input)
	assert.NotNil(t, err)
}
