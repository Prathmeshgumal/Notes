package tui

import (
	"regexp"
	"strings"
)

var (
	// A top-level bullet or task item: "- x", "* x", "+ x", "- [ ] x".
	topLevelBullet = regexp.MustCompile(`^[-*+][ \t]`)
	// Any list item, at any indent, ordered or not.
	anyListItem = regexp.MustCompile(`^[ \t]*([-*+]|\d+[.)])[ \t]`)
	// Ordered items are left alone: splitting one restarts its numbering.
	orderedItem   = regexp.MustCompile(`^[ \t]*\d+[.)][ \t]`)
	codeFence     = regexp.MustCompile("^[ \t]*(```|~~~)")
	listSeparator = "<!-- -->"
)

// A paragraph holding only a zero-width space renders as an empty line, which
// is how a run of blank lines is reproduced. The renderer puts one blank line
// above and below every block, so each filler is worth two blank lines and a
// run comes out as 1+2n — even-length runs land on the nearest odd number below.
const zeroWidthSpace = "\u200b"

// fillersFor converts the length of a blank run into the number of filler
// paragraphs needed to reproduce roughly that much vertical space.
func fillersFor(blankRun int) int {
	if blankRun < 2 {
		return 0
	}
	return (blankRun - 1) / 2
}

// separateListGroups keeps the blank lines a writer puts between groups of
// bullets visible in the rendered output.
//
// Markdown treats a blank line inside a list as part of the same list, and
// collapses any run of blank lines into a single break, so the renderer closes
// the gap. Emitting an empty HTML comment ends one list and starts another,
// and filler paragraphs after it reproduce the height of the original run.
// Ordered lists are skipped, since splitting them restarts numbering.
func separateListGroups(md string) string {
	if !strings.Contains(md, "\n\n") {
		return md
	}
	lines := strings.Split(md, "\n")
	out := make([]string, 0, len(lines)+4)

	inFence := false
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if codeFence.MatchString(line) {
			inFence = !inFence
		}
		out = append(out, line)
		if inFence || strings.TrimSpace(line) != "" {
			continue
		}

		// A run of blank lines: decide whether it separates two bullet groups.
		prev := lastNonBlank(lines[:i])
		j := i
		for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
			j++
		}
		if j >= len(lines) {
			continue
		}
		next := lines[j]

		if anyListItem.MatchString(prev) && !orderedItem.MatchString(prev) &&
			topLevelBullet.MatchString(next) {
			// One blank line is already in `out`; replace the rest of the run
			// with a list break plus enough fillers to match its height.
			out = append(out, listSeparator, "")
			for k := fillersFor(j - i); k > 0; k-- {
				out = append(out, zeroWidthSpace, "")
			}
			i = j - 1 // resume at the line that ended the run
		}
	}
	return strings.Join(out, "\n")
}

func lastNonBlank(lines []string) string {
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return lines[i]
		}
	}
	return ""
}

// stripDerivedTitle removes the opening line of a note when it is the line the
// title was taken from, so the preview does not show the same text as both its
// header and its first line. A title the writer typed explicitly leaves the
// body untouched.
func stripDerivedTitle(content, title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return content
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue // leading blank lines
		}
		heading := strings.TrimSpace(strings.TrimLeft(line, "#"))
		if heading != title {
			return content
		}
		// Drop the line, plus the blank lines that followed it, so the body
		// does not start with a gap where the heading used to be.
		rest := lines[i+1:]
		for len(rest) > 0 && strings.TrimSpace(rest[0]) == "" {
			rest = rest[1:]
		}
		return strings.Join(rest, "\n")
	}
	return content
}

// countTasks reports how many task items a note has and how many are ticked,
// for the details shown beside it.
func countTasks(content string) (done, total int) {
	for _, line := range strings.Split(content, "\n") {
		m := taskPrefix.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		total++
		if strings.ContainsAny(m[0], "xX") {
			done++
		}
	}
	return
}
