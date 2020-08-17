package ink

type Section struct {
	text string
	tags []string

	opts     []string
	optsTags [][]string

	end bool
}

func (s *Section) add(text string, tags []string) {
	if len(text) > 0 {
		s.text = s.text + "\n" + text
	}

	s.tags = append(s.tags, tags...)
}

type Nodes []Node

func (n Nodes) NewSection() *Section {
	sec := &Section{}
	for _, node := range n {
		switch node.(type) {
		case End:
			sec.end = true
			sec.add(node.(End).End())
		case Choices:
			opts, optsTags := node.(Choices).List()
			sec.opts = opts
			sec.optsTags = optsTags
		case CanNext:
			sec.add(node.(CanNext).Render())
		}
	}
	return sec
}
