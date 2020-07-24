package goink

import (
	"regexp"
)

// Inline content
type Inline struct {
	parent  Block
	content []Block

	raw string
}

// Parent of the inline block
func (i *Inline) Parent() Block {
	return i.parent
}

// Content of the inline block
func (i *Inline) Content() []Block {
	return i.content
}

var (
	divertReg = regexp.MustCompile(`\s->\s(\w+)$`)
)

func (i *Inline) parse(raw string) (blk Block, err error) {
	// INLINE
	i.raw = raw

	// -> divert
	var divert *Divert
	if d := divertReg.FindStringSubmatch(raw); d != nil {
		divert = &Divert{link: d[1], parent: i}
	}

	// add divert into content at last
	if divert != nil {
		i.content = append(i.content, divert)
	}

	// inline content will always return nothing
	return nil, nil
}

// newInline generator
func newInline(str string) *Inline {
	i := &Inline{}
	i.parse(str)
	return i
}

// Text of the ink
type Text struct {
	str    string
	parent Block
}

// Parent of the inline text
func (t *Text) Parent() Block {
	return t.parent
}

// Content of the inline block
func (t *Text) Content() []Block {
	return nil
}

func (t *Text) parse(raw string) (blk Block, err error) {
	return nil, nil
}

// Divert of the ink
type Divert struct {
	link   string
	parent Block
}

// Parent of the inline block
func (d *Divert) Parent() Block {
	return d.parent
}

// Content of the inline block
func (d *Divert) Content() []Block {
	// knot := d.Parent()
	// ok := false
	// for {
	// 	knot, ok = knot.(*Knot)
	// 	if !ok {
	// 		knot = knot.Parent()
	// 	} else {
	// 		break
	// 	}
	// }

	// // TODO: stitch
	// // TODO: sub stich

	// // Find target knot, and return its content
	// if target := knot.(*Knot).story.findKnot(d.link); target != nil {
	// 	return target.Content()
	// }

	return nil
}

func (d *Divert) parse(raw string) (blk Block, err error) {
	return nil, nil
}
