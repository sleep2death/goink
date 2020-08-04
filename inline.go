package goink

// Basic Node of the story
type InkObj interface {
	Story() *Story
	Parent() InkObj

	Next() InkObj
	SetNext(obj InkObj)
}

func NewInline(s *Story, input string) error {
	// Inline
	i := &Inline{story: s, parent: s.current, raw: input}
	s.current.SetNext(i)
	s.current = i
	return nil
}

// Inline node of tye story
type Inline struct {
	story  *Story
	parent InkObj
	next   InkObj

	raw string
}

func (i *Inline) Render() string {
	return i.raw
}

func (i *Inline) Parent() InkObj {
	return i.parent
}

func (i *Inline) SetNext(obj InkObj) {
	i.next = obj
}

func (i *Inline) Next() InkObj {
	if i.next != nil {
		return i.next
	}

	obj := i.parent

	for obj != nil {
		if c, ok := obj.(*Choices); ok {
			if c.gather != nil {
				return c.gather
			}
		}

		obj = obj.Parent()
	}

	return nil
}

func (i *Inline) Story() *Story {
	return i.story
}
