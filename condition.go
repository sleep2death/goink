package goink

import (
	"regexp"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/pkg/errors"
)

var regReplaceDot = regexp.MustCompile(`\.(\w+)`)

// exprc in the line
type exprc struct {
	// env     map[string]interface{}
	program *vm.Program
	raw     string
}

// NewExprc creates a condition with the given expr
func NewExprc(code string) (*exprc, error) {
	cond := &exprc{raw: code}
	c := regReplaceDot.ReplaceAllString(code, SPLIT+"$1")

	program, err := expr.Compile(c, expr.Env(nil))

	if err != nil {
		return nil, err
	}

	cond.program = program
	return cond, nil
}

// Bool return the exprc result as bool value
func (c *exprc) Bool(count map[string]int) (bool, error) {
	output, err := expr.Run(c.program, count)
	if err != nil {
		return false, err
	}

	// fmt.Println(c.program.Source.Content(), output, count["Knot_A-gather"])

	b, ok := output.(bool)
	if ok {
		return b, nil
	}

	i, ok := output.(int)
	if ok {
		return (i > 0), nil
	}

	return false, errors.Errorf("output is not a bool value: %v", output)
}
