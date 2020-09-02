package parser

import (
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// position describes an arbitrary source position
// including the file, line, and column location.
// A Position is valid if the line number is > 0.
//
type position struct {
	offset int // offset, starting at 0
	line   int // line number, starting at 1
	column int // column number, starting at 1 (byte count)
}

// IsValid reports whether the position is valid.
func (pos *position) IsValid() bool { return pos.line > 0 }

// string returns a string in one of several forms:
//
//	line:column         valid position without file name
//	line                valid position without file name and no column (column == 0)
//	file                invalid position with file name
//	-                   invalid position without file name
//
func (pos position) String() string {
	var s string
	if pos.IsValid() {
		s += fmt.Sprintf("%d", pos.line)
		if pos.column != 0 {
			s += fmt.Sprintf(":%d", pos.column)
		}
	}
	if s == "" {
		s = "-"
	}
	return s
}

// pos is a compact encoding of a source position within a file set.
// It can be converted into a Position for a more convenient, but much
// larger, representation.
//
type pos int

// The zero value for pos is NoPos; there is no file and line information
// associated with it, and noPos.IsValid() is false. noPos is always
// smaller than any other pos value. The corresponding position value
// for noPos is the zero value for position.
//
const noPos pos = 0

// IsValid reports whether the position is valid.
func (p pos) IsValid() bool {
	return p != noPos
}

type errorHandler func(pos position, msg string)

type scanner struct {
	// immutable state
	size  int // source file handle
	lines []int

	dir string       // directory portion of file.Name()
	src []byte       // source
	err errorHandler // error reporting; or nil

	// scanning state
	ch         rune // current character
	offset     int  // character offset
	rdOffset   int  // reading offset (position after current character)
	lineOffset int  // current line offset
	insertSemi bool // insert a semicolon before next newline

	// public state - ok to modify
	ErrorCount int // number of errors encountered
}

const bom = 0xFEFF // byte order mark, only permitted as very first character

// Read the next Unicode char into s.ch.
// s.ch < 0 means end-of-file.
//
func (s *scanner) next() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.addLine(s.offset)
		}
		r, w := rune(s.src[s.rdOffset]), 1
		switch {
		case r == 0:
			s.error(s.offset, "illegal character NUL")
		case r >= utf8.RuneSelf:
			// not ASCII
			r, w = utf8.DecodeRune(s.src[s.rdOffset:])
			if r == utf8.RuneError && w == 1 {
				s.error(s.offset, "illegal UTF-8 encoding")
			} else if r == bom && s.offset > 0 {
				s.error(s.offset, "illegal byte order mark")
			}
		}
		s.rdOffset += w
		s.ch = r
	} else {
		s.offset = len(s.src)
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.addLine(s.offset)
		}
		s.ch = -1 // eof
	}
}

// peek returns the byte following the most recently read character without
// advancing the scanner. If the scanner is at EOF, peek returns 0.
//
func (s *scanner) peek() byte {
	if s.rdOffset < len(s.src) {
		return s.src[s.rdOffset]
	}
	return 0
}

// init prepares the scanner s to tokenize the text src by setting the
// scanner at the beginning of src.
//
func (s *scanner) init(src []byte, err errorHandler) {
	// Explicitly initialize all fields since a scanner may be reused.
	s.size = len(src)
	s.src = src
	s.err = err

	s.ch = ' '
	s.offset = 0
	s.rdOffset = 0
	s.lineOffset = 0
	s.insertSemi = false
	s.ErrorCount = 0

	s.next()
	if s.ch == bom {
		s.next() // ignore BOM at file beginning
	}
}

func (s *scanner) error(offs int, msg string) {
	if s.err != nil {
		s.err(s.position(s.pos(offs)), msg)
	}
	s.ErrorCount++
}

func (s *scanner) errorf(offs int, format string, args ...interface{}) {
	s.error(offs, fmt.Sprintf(format, args...))
}

func (s *scanner) scanComment() string {
	// initial '/' already consumed; s.ch == '/' || s.ch == '*'
	offs := s.offset - 1 // position of initial '/'
	next := -1           // position immediately following the comment; < 0 means invalid comment
	numCR := 0

	if s.ch == '/' {
		//-style comment
		// (the final '\n' is not considered part of the comment)
		s.next()
		for s.ch != '\n' && s.ch >= 0 {
			if s.ch == '\r' {
				numCR++
			}
			s.next()
		}
		// if we are at '\n', the position following the comment is afterwards
		next = s.offset
		if s.ch == '\n' {
			next++
		}
		goto exit
	}

	/*-style comment */
	s.next()
	for s.ch >= 0 {
		ch := s.ch
		if ch == '\r' {
			numCR++
		}
		s.next()
		if ch == '*' && s.ch == '/' {
			s.next()
			next = s.offset
			goto exit
		}
	}

	s.error(offs, "comment not terminated")

exit:
	lit := s.src[offs:s.offset]

	// On Windows, a (//-comment) line may end in "\r\n".
	// Remove the final '\r' before analyzing the text for
	// line directives (matching the compiler). Remove any
	// other '\r' afterwards (matching the pre-existing be-
	// havior of the scanner).
	if numCR > 0 && len(lit) >= 2 && lit[1] == '/' && lit[len(lit)-1] == '\r' {
		lit = lit[:len(lit)-1]
		numCR--
	}

	// interpret line directives
	// (//line directives must start at the beginning of the current line)
	/* if next >= 0 && (lit[1] == '*' || offs == s.lineOffset) && bytes.HasPrefix(lit[2:], prefix) {
		s.updateLineInfo(next, offs, lit)
	} */

	if numCR > 0 {
		lit = stripCR(lit, lit[1] == '*')
	}

	return string(lit)
}

func stripCR(b []byte, comment bool) []byte {
	c := make([]byte, len(b))
	i := 0
	for j, ch := range b {
		// In a /*-style comment, don't strip \r from *\r/ (incl.
		// sequences of \r from *\r\r...\r/) since the resulting
		// */ would terminate the comment too early unless the \r
		// is immediately following the opening /* in which case
		// it's ok because /*/ is not closed yet (issue #11151).
		if ch != '\r' || comment && i > len("/*") && c[i-1] == '*' && j+1 < len(b) && b[j+1] == '/' {
			c[i] = ch
			i++
		}
	}
	return c[:i]
}

func (s *scanner) scanString() string {
	offs := s.offset
	numCR := 0
	for !isLineBreaker(s.ch) && s.ch >= 0 {
		if s.ch == '\r' {
			numCR++
		}
		s.next()
	}

	lit := s.src[offs:s.offset]

	// On Windows, a (//-comment) line may end in "\r\n".
	// Remove the final '\r' before analyzing the text for
	// line directives (matching the compiler). Remove any
	// other '\r' afterwards (matching the pre-existing be-
	// havior of the scanner).
	if numCR > 0 && lit[len(lit)-1] == '\r' {
		lit = lit[:len(lit)-1]
		numCR--
	}
	return string(lit)
}

func (s *scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' && !s.insertSemi || s.ch == '\r' {
		s.next()
	}
}

// Scan scans the next token and returns the token position, the token,
// and its literal string if applicable. The source end is indicated by
// token.EOF.
//
func (s *scanner) scan() (pos pos, tok token, lit string) {
	s.skipWhitespace()

	// current token start
	pos = s.pos(s.offset)

	ch := s.ch
	s.next()

	switch ch {
	case -1:
		if s.insertSemi {
			s.insertSemi = false // EOF consumed
			return pos, SEMICOLON, "\n"
		}
		tok = EOF
	case '\n':
		// we only reach here if s.insertSemi was
		// set in the first place and exited early
		// from s.skipWhitespace()
		s.insertSemi = false // newline consumed
		return pos, SEMICOLON, "\n"
	case '/':
		if s.ch == '/' || s.ch == '*' {
			// comment
			if s.insertSemi && s.findLineEnd() {
				// reset position to the beginning of the comment
				s.ch = '/'
				s.offset = int(pos)
				s.rdOffset = s.offset + 1
				s.insertSemi = false // newline consumed
				return pos, SEMICOLON, "\n"
			}
			comment := s.scanComment()
			tok = COMMENT
			lit = comment
		} else {
			tok = CHAR
			lit = string(ch)
		}
	case '#': // tag
	case '{': // expr
	case '(': // label
	case '[': // knot
	case '>': // divert
	case '-': // gather
	case '+': // sticky option
	case '*': // once-only option
	default:
		lit = string(ch) + s.scanString()
		tok = STRING
	}

	return
}

func isLineBreaker(ch rune) bool {
	// tag | comment | divert | expr | eof
	return ch == '#' || ch == '/' || ch == '>' || ch == '{' || ch == -1
}

func (s *scanner) findLineEnd() bool {
	// initial '/' already consumed

	defer func(offs int) {
		// reset scanner state to where it was upon calling findLineEnd
		s.ch = '/'
		s.offset = offs
		s.rdOffset = offs + 1
		s.next() // consume initial '/' again
	}(s.offset - 1)

	// read ahead until a newline, EOF, or non-comment token is found
	for s.ch == '/' || s.ch == '*' {
		if s.ch == '/' {
			//-style comment always contains a newline
			return true
		}
		/*-style comment: look for newline */
		s.next()
		for s.ch >= 0 {
			ch := s.ch
			if ch == '\n' {
				return true
			}
			s.next()
			if ch == '*' && s.ch == '/' {
				s.next()
				break
			}
		}
		s.skipWhitespace() // s.insertSemi is set
		if s.ch < 0 || s.ch == '\n' {
			return true
		}
		if s.ch != '/' {
			// non-comment token
			return false
		}
		s.next() // consume '/'
	}

	return false
}

func isLetter(ch rune) bool {
	return 'a' <= lower(ch) && lower(ch) <= 'z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return isDecimal(ch) || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

func (s *scanner) scanIdentifier() string {
	offs := s.offset
	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}
	return string(s.src[offs:s.offset])
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= lower(ch) && lower(ch) <= 'f':
		return int(lower(ch) - 'a' + 10)
	}
	return 16 // larger than any legal digit val
}

func lower(ch rune) rune     { return ('a' - 'A') | ch } // returns lower-case ch iff ch is ASCII letter
func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }
func isHex(ch rune) bool     { return '0' <= ch && ch <= '9' || 'a' <= lower(ch) && lower(ch) <= 'f' }

func (s *scanner) addLine(offset int) {
	if i := len(s.lines); (i == 0 || s.lines[i-1] < offset) && offset < s.size {
		s.lines = append(s.lines, offset)
	}
}

// unpack returns the filename and line and column number for a file offset.
func (s *scanner) unpack(offset int) (line, column int) {
	if i := searchInts(s.lines, offset); i >= 0 {
		line, column = i+1, offset-s.lines[i]+1
	}
	return
}

//
func (s *scanner) pos(offset int) pos {
	if offset > s.size {
		panic("illegal file offset")
	}
	return pos(offset)
}

func (s *scanner) position(p pos) (pos position) {
	offset := int(p)
	pos.offset = offset
	pos.line, pos.column = s.unpack(offset)
	return
}

// -----------------------------------------------------------------------------
// Helper functions

func searchInts(a []int, x int) int {
	i, j := 0, len(a)
	for i < j {
		h := i + (j-i)/2 // avoid overflow when computing h
		// i ≤ h < j
		if a[h] <= x {
			i = h + 1
		} else {
			j = h
		}
	}
	return i - 1
}

// -----------------------------------------------------------------------------
// Token is the set of lexical tokens of the goink.
type token int

// The list of tokens.
const (
	// Special tokens
	ILLEGAL token = iota
	EOF
	COMMENT

	SEMICOLON // ;
	STRING
	CHAR

	TAG   // #
	EXPR  // {}
	LABEL // ()

	KNOT   // ==
	DIVERT // ->
	OPTION // *
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	COMMENT: "COMMENT",

	SEMICOLON: "SEMICOLON",
	STRING:    "STRING",
	CHAR:      "CHAR",

	TAG:   "TAG",
	EXPR:  "EXPR",
	LABEL: "LABEL",

	KNOT:   "KNOT",
	DIVERT: "DIVERT",
	OPTION: "CHOICE",
}

func (tok token) String() string {
	s := ""
	if 0 <= tok && tok < token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}
