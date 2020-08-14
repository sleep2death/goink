package goink

// Section of the story
type Section struct {
	text    string
	tags    []string
	options []*opt
	end     bool
}
