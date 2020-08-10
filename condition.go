package goink

import (
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type Condition struct {
	// env     map[string]interface{}
	program *vm.Program
	raw     string
}

type env struct {
	count map[string]int
	vars  map[string]interface{}
}

func NewCondition(code string) *Condition {
	cond := &Condition{raw: code}
	program, err := expr.Compile(code, expr.Env(nil))

	if err != nil {
		panic(err)
	}

	cond.program = program
	return cond
}
