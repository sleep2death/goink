package goink

type Condition struct {
	raw string
}

func NewCondition(input string) *Condition {
	return &Condition{raw: input}
}
