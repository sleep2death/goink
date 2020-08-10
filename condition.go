package goink

import (
	"fmt"
	"github.com/antonmedv/expr"
)

type Condition struct {
	env map[string]interface{}
	raw string
}

func LargerThanZero(input int) bool {
	return input > 0
}

func NewCondition(code string) *Condition {
	cond := &Condition{raw: code}

	env := map[string]interface{}{
		"greet":          "Hello, %v!",
		"conditional_a":  4,
		"conditional_b":  0,
		"names":          []string{"world", "you"},
		"largerThanZero": LargerThanZero,
	}

	cond.env = env

	program, err := expr.Compile(code, expr.Env(nil))

	if err != nil {
		panic(err)
	}

	output, err := expr.Run(program, env)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(output)
	}

	return cond
}
