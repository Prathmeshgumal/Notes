package tui

import "strings"

// Cursor arithmetic for the editor. The textarea works in rows and columns;
// these translate to and from an offset into the whole note so the formatting
// helpers can treat it as one string.

func offsetAt(value string, row, col int) int {
	lines := strings.Split(value, "\n")
	if row < 0 {
		row = 0
	}
	if row >= len(lines) {
		row = len(lines) - 1
	}
	off := 0
	for i := 0; i < row; i++ {
		off += len(lines[i]) + 1 // the newline
	}
	if col > len(lines[row]) {
		col = len(lines[row])
	}
	if col < 0 {
		col = 0
	}
	return off + col
}

func rowColAt(value string, offset int) (int, int) {
	if offset < 0 {
		offset = 0
	}
	if offset > len(value) {
		offset = len(value)
	}
	row := strings.Count(value[:offset], "\n")
	lineStart := strings.LastIndex(value[:offset], "\n") + 1
	return row, offset - lineStart
}

// wordAt returns the bounds of the run of non-space characters around offset,
// with trailing sentence punctuation left outside so "word." wraps as "word".
func wordAt(value string, offset int) (int, int) {
	if offset > len(value) {
		offset = len(value)
	}
	isSpace := func(b byte) bool { return b == ' ' || b == '\t' || b == '\n' }

	start := offset
	for start > 0 && !isSpace(value[start-1]) {
		start--
	}
	end := offset
	for end < len(value) && !isSpace(value[end]) {
		end++
	}
	for end > start && strings.IndexByte(".,;:!?", value[end-1]) >= 0 {
		end--
	}
	return start, end
}

// mark wraps the word under the cursor in open/close, unwraps it if it is
// already wrapped, and inserts an empty pair when the cursor is not on a word.
// It returns the new text and where the cursor should land.
func mark(value string, offset int, open, close string) (string, int) {
	start, end := wordAt(value, offset)

	if start == end {
		return value[:offset] + open + close + value[offset:], offset + len(open)
	}

	// Already wrapped? Take the marks back off. They sit inside the run, since
	// "*" is not whitespace.
	if inner := value[start:end]; wrappedWith(inner, open, close) {
		stripped := inner[len(open) : len(inner)-len(close)]
		return value[:start] + stripped + value[end:], start + len(stripped)
	}

	return value[:start] + open + value[start:end] + close + value[end:],
		end + len(open) + len(close)
}

// wrappedWith reports whether text already carries these marks. A single "*"
// must not match text that is really bold, so a second pair inside means the
// marks belong to a different level of emphasis.
func wrappedWith(text, open, close string) bool {
	if len(text) < len(open)+len(close) {
		return false
	}
	if !strings.HasPrefix(text, open) || !strings.HasSuffix(text, close) {
		return false
	}
	inner := text[len(open) : len(text)-len(close)]
	if strings.HasPrefix(inner, open) && strings.HasSuffix(inner, close) {
		return false
	}
	return true
}

// link turns the word under the cursor into [word](), leaving the cursor
// between the brackets so the address can be typed straight in.
func link(value string, offset int) (string, int) {
	start, end := wordAt(value, offset)
	if start == end {
		return value[:offset] + "[]()" + value[offset:], offset + 1
	}
	word := value[start:end]
	return value[:start] + "[" + word + "]()" + value[end:],
		start + len(word) + 3
}
