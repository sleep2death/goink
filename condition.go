package goink

import (
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/pkg/errors"
)

// Condition of the inline
type Condition struct {
	// env     map[string]interface{}
	program *vm.Program
	raw     string
}

// NewCondition creates a condition with the given expr
func NewCondition(code string) (*Condition, error) {
	cond := &Condition{raw: code}
	program, err := expr.Compile(code, expr.Env(nil))

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

	b, ok := output.(bool)
	if !ok {
		return false, errors.Errorf("output is not a bool value: %s", c.program.Source.Content())
	}

	return b, nil

}
