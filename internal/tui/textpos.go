package tui

import (
	"regexp"
	"strings"
)

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

// Line-based formatting. Each takes the whole note and a cursor offset, and
// returns the new text with the cursor kept on the same line.

func lineBounds(value string, offset int) (int, int) {
	if offset > len(value) {
		offset = len(value)
	}
	start := strings.LastIndexByte(value[:offset], '\n') + 1
	end := strings.IndexByte(value[start:], '\n')
	if end < 0 {
		return start, len(value)
	}
	return start, start + end
}

// editLine replaces the line under the cursor with fn(line), leaving the cursor
// at the same position within it.
func editLine(value string, offset int, fn func(string) string) (string, int) {
	start, end := lineBounds(value, offset)
	old := value[start:end]
	next := fn(old)
	return value[:start] + next + value[end:], offset + len(next) - len(old)
}

// indentOf splits a line into its leading whitespace and the rest, so a marker
// added to an indented line goes after the indentation, not before it.
func indentOf(line string) (string, string) {
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	return line[:i], line[i:]
}

var (
	headingPrefix = regexp.MustCompile(`^(#{1,6}) `)
	bulletPrefix  = regexp.MustCompile(`^(\s*)[-*+] `)
	numberPrefix  = regexp.MustCompile(`^(\s*)\d+\. `)
	taskPrefix    = regexp.MustCompile(`^(\s*)[-*+] \[[ xX]\] `)
	quotePrefix   = regexp.MustCompile(`^> ?`)
)

// heading cycles the line through #, ##, … ###### and back to plain, the way
// the web toolbar's heading button does.
func heading(value string, offset int) (string, int) {
	return editLine(value, offset, func(line string) string {
		m := headingPrefix.FindStringSubmatch(line)
		if m == nil {
			return "# " + line
		}
		if len(m[1]) >= 6 {
			return line[len(m[0]):]
		}
		return "#" + line
	})
}

// bullet toggles "- " on the line, keeping any indentation.
func bullet(value string, offset int) (string, int) {
	return editLine(value, offset, func(line string) string {
		if m := taskPrefix.FindStringSubmatch(line); m != nil {
			return m[1] + "- " + line[len(m[0]):] // a task is already a bullet
		}
		if m := bulletPrefix.FindStringSubmatch(line); m != nil {
			return m[1] + line[len(m[0]):]
		}
		if m := numberPrefix.FindStringSubmatch(line); m != nil {
			return m[1] + "- " + line[len(m[0]):]
		}
		indent, rest := indentOf(line)
		return indent + "- " + rest
	})
}

// numbered toggles "1. " on the line. The number is left at 1: Markdown
// renumbers a list from its first item, so the rendered output is still right.
func numbered(value string, offset int) (string, int) {
	return editLine(value, offset, func(line string) string {
		if m := numberPrefix.FindStringSubmatch(line); m != nil {
			return m[1] + line[len(m[0]):]
		}
		if m := taskPrefix.FindStringSubmatch(line); m != nil {
			return m[1] + "1. " + line[len(m[0]):]
		}
		if m := bulletPrefix.FindStringSubmatch(line); m != nil {
			return m[1] + "1. " + line[len(m[0]):]
		}
		indent, rest := indentOf(line)
		return indent + "1. " + rest
	})
}

// task toggles "- [ ] ", and ticks or unticks a box that is already there.
func task(value string, offset int) (string, int) {
	return editLine(value, offset, func(line string) string {
		if m := taskPrefix.FindStringSubmatch(line); m != nil {
			return m[1] + "- " + line[len(m[0]):]
		}
		if m := bulletPrefix.FindStringSubmatch(line); m != nil {
			return m[1] + "- [ ] " + line[len(m[0]):]
		}
		if m := numberPrefix.FindStringSubmatch(line); m != nil {
			return m[1] + "- [ ] " + line[len(m[0]):]
		}
		indent, rest := indentOf(line)
		return indent + "- [ ] " + rest
	})
}

// toggleTick flips a task between done and not done, which is what a reader
// actually wants from a checklist.
func toggleTick(value string, offset int) (string, int) {
	return editLine(value, offset, func(line string) string {
		m := taskPrefix.FindStringSubmatch(line)
		if m == nil {
			return line
		}
		box := m[0][len(m[1]) : len(m[0])-1] // "- [ ]" or "- [x]"
		if strings.ContainsAny(box, "xX") {
			return m[1] + "- [ ] " + line[len(m[0]):]
		}
		return m[1] + "- [x] " + line[len(m[0]):]
	})
}

func quote(value string, offset int) (string, int) {
	return editLine(value, offset, func(line string) string {
		if m := quotePrefix.FindString(line); m != "" {
			return line[len(m):]
		}
		return "> " + line
	})
}

// rule inserts a horizontal rule on its own line below the cursor.
func rule(value string, offset int) (string, int) {
	_, end := lineBounds(value, offset)
	insert := "\n\n---\n"
	return value[:end] + insert + value[end:], end + len(insert)
}

// codeBlock wraps the current line in a fence, or inserts an empty one.
func codeBlock(value string, offset int) (string, int) {
	start, end := lineBounds(value, offset)
	line := value[start:end]
	if strings.TrimSpace(line) == "" {
		insert := "```\n\n```"
		return value[:start] + insert + value[end:], start + 4
	}
	return value[:start] + "```\n" + line + "\n```" + value[end:], end + 4
}
