package goink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGather(t *testing.T) {
	input := `
	Hello World,
	* Option A
	* Option B
	- --> Knot_A
	# Knot_A
	A content here.
	`
	_, err := parse(input)
	assert.NotNil(t, err)
}
