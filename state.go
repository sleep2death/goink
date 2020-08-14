package goink

// NewState of the current story
func NewState(story *Story) *State {
	path := story.c.Path()
	count := make(map[string]int)

	// copy all non-zero count into state's count
	for k, v := range story.vars {
		if v > 0 {
			count[k] = v
		}
	}

	return &State{path: path, count: count}
}

// State of the current story
type State struct {
	path  string
	count map[string]int
}

// Path of the current inkObj
func (s *State) Path() string {
	return s.path
}

// Count collection of the current story
func (s *State) Count() map[string]int {
	return s.count
}
