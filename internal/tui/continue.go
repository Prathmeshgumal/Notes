package tui

import (
	"regexp"
	"strconv"
	"strings"
)

// Pressing Enter inside a list carries the list on to the next line, the way a
// notes app is expected to. Pressing it on an item with nothing in it ends the
// list instead, which is how a list is finished without reaching for a mouse.

var (
	// "- ", "* ", "+ ", with any indent, optionally followed by a checkbox.
	bulletLine = regexp.MustCompile(`^(\s*)([-*+])(\s+)(\[[ xX]\]\s+)?(.*)$`)
	// "1. ", "2) ", with any indent.
	orderedLine = regexp.MustCompile(`^(\s*)(\d+)([.)])(\s+)(.*)$`)
	// "> ", nested quotes included.
	quoteLine = regexp.MustCompile(`^(\s*)((?:>\s*)+)(.*)$`)
)

// continuation reports what a new line should begin with when Enter is pressed
// on the given line.
//
// prefix is what to type after the line break. endList is true when the line is
// an empty item, meaning the marker should be cleared rather than repeated —
// the writer is telling us the list is over.
func continuation(line string) (prefix string, endList bool) {
	if m := bulletLine.FindStringSubmatch(line); m != nil {
		indent, marker, gap, box, text := m[1], m[2], m[3], m[4], m[5]
		if strings.TrimSpace(text) == "" {
			return "", true
		}
		if box != "" {
			// A new task starts unticked however the one above it ended.
			return indent + marker + gap + "[ ] ", false
		}
		return indent + marker + gap, false
	}

	if m := orderedLine.FindStringSubmatch(line); m != nil {
		indent, num, dot, gap, text := m[1], m[2], m[3], m[4], m[5]
		if strings.TrimSpace(text) == "" {
			return "", true
		}
		n, err := strconv.Atoi(num)
		if err != nil {
			return "", false
		}
		return indent + strconv.Itoa(n+1) + dot + gap, false
	}

	if m := quoteLine.FindStringSubmatch(line); m != nil {
		indent, markers, text := m[1], m[2], m[3]
		if strings.TrimSpace(text) == "" {
			return "", true
		}
		return indent + markers, false
	}

	return "", false
}
