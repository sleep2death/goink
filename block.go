package goink

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"
)

// ReadLines from ink file.
func readInk(path string) (b *block, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	reader := bufio.NewReader(file)
	buffer := bytes.NewBuffer(make([]byte, 1024))

	b = &block{}
	for {
		part, prefix, err := reader.ReadLine()
		if err != nil {
			break
		}

		buffer.Write(part)

		if !prefix {
			b, err = b.parse(strings.TrimRight(strings.TrimSpace(buffer.String()), "\r\n"))
			if err != nil {
				return b, err
			}

			buffer.Reset()
		}
	}

	if err == io.EOF {
		err = nil
	}
	return
}

type blockType int

const (
	blkKnot blockType = iota
	blkChoice
	blkStitch
	blkInline
)

type block struct {
	parent  *block
	content []*block
	inline  *inline

	bt     blockType
	nested int
}

func (b *block) isRoot() bool {
	return b.parent == nil
}

var blkReg = regexp.MustCompile(`(^\={2,}\s)|(^\++\s)`)

func (b *block) root() *block {
	var root *block = b
	for {
		if root != nil && root.parent != nil {
			root = root.parent
		} else {
			break
		}
	}

	return root
}

func (b *block) parse(input string) (*block, error) {
	// skip empty line
	if len(input) == 0 {
		return b, nil
	}
	// find block header
	res := blkReg.FindStringSubmatch(input)
	blk := &block{inline: &inline{raw: input}}
	blk.inline.parse()

	if len(res) > 0 { // found block header
		if len(res[1]) > 0 { // KNOT
			blk.bt = blkKnot
			blk.nested = 1 // knot's nest level must be 1

			root := b.root() // finding root block
			blk.parent = root
			root.content = append(root.content, blk)
			// TODO: parse knot header
		} else if len(res[2]) > 0 { // CHOICE
			blk.bt = blkChoice
			blk.nested = len(res[2]) - 1 // root choice level is 1

			if b.bt == blkKnot { // prev block is knot
				blk.parent = b
				b.content = append(b.content, blk)
			} else if b.bt == blkChoice {
				for b.nested >= blk.nested && b.bt != blkKnot {
					b = b.parent
				}
				blk.parent = b
				b.content = append(b.content, blk)
			}
		}

		// fmt.Println(blk.content, blk.nested, "<", blk.parent.content)
		return blk, nil // always return blk as following container
		// return nil, nil
	}

	// found inline block
	blk.bt = blkInline
	blk.nested = b.nested

	blk.parent = b
	b.content = append(b.content, blk)
	// fmt.Println(blk.content, "<", blk.parent.content)
	return b, nil
}

func (b *block) format(indent string) (res string) {
	if b.inline != nil && len(b.inline.raw) > 0 {
		res += indent + b.inline.raw + "\n"
	}

	if b.parent != nil {
		indent += "  "
	}
	for _, blk := range b.content {
		res += blk.format(indent)
	}
	return
}

type content interface {
}

type inline struct {
	raw      string
	children []content
}

func (il *inline) parse() {
	content := newKnotHeader(il.raw)
	if content != nil {
		il.children = append(il.children, content)
	}
}

type knotHeader struct {
	name string
}

var knotReg = regexp.MustCompile(`^={2,}\s(\w+)`)

func newKnotHeader(str string) content {
	res := knotReg.FindStringSubmatch(str)
	if len(res[0]) > 0 {
		return &knotHeader{name: res[1]}
	}
	return nil
}
