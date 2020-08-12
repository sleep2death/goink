package goink

import (
	"regexp"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/pkg/errors"
)

var regReplaceDot = regexp.MustCompile(`\.(\w+)`)

// Condition of the inline
type Condition struct {
	// env     map[string]interface{}
	program *vm.Program
	raw     string
}

// NewCondition creates a condition with the given expr
func NewCondition(code string) (*Condition, error) {
	cond := &Condition{raw: code}
	c := regReplaceDot.ReplaceAllString(code, split+"$1")

	program, err := expr.Compile(c, expr.Env(nil))

	if err != nil {
		return nil, err
	}

	cond.program = program
	return cond, nil
}

// Bool return the expr result as bool value
func (c *Condition) Bool(count map[string]int) (bool, error) {
	output, err := expr.Run(c.program, count)
	if err != nil {
		return false, err
	}

	// fmt.Println(c.program.Source.Content(), output, count["Knot_A-gather"])

	b, ok := output.(bool)
	if !ok {
		return false, errors.Errorf("output is not a bool value: %s", c.program.Source.Content())
	}

	return b, nil

}
