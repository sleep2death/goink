package goink

// Section of the story
type Section struct {
	text     string
	tags     []string
	opts     []string
	optsTags [][]string
	end      bool
}
