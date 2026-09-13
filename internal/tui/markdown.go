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

// separateListGroups keeps the blank lines a writer puts between groups of
// bullets visible in the rendered output.
//
// Markdown treats a blank line inside a list as part of the same list, so the
// renderer closes the gap. Emitting an empty HTML comment between the groups
// ends one list and starts another, which restores the spacing without showing
// anything. Ordered lists are skipped, since splitting them restarts numbering.
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
			// Copy the rest of the blank run, then break the list.
			for ; i+1 < j; i++ {
				out = append(out, lines[i+1])
			}
			out = append(out, listSeparator, "")
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
