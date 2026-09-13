package tui

import (
	"regexp"
	"strings"
	"testing"
)

var ansiCodes = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// renderedBlankLines counts the blank lines glamour emits between the first
// and second list group, which is the spacing the reader actually sees.
func renderedBlankLines(t *testing.T, md string) int {
	t.Helper()
	r, err := newRenderer("dark", 60)
	if err != nil {
		t.Fatal(err)
	}
	out, err := r.Render(md)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.Trim(ansiCodes.ReplaceAllString(out, ""), "\n"), "\n")

	first, second := -1, -1
	for i, l := range lines {
		if strings.Contains(l, "first group") && first == -1 {
			first = i
		}
		if strings.Contains(l, "second group") {
			second = i
		}
	}
	if first < 0 || second < 0 {
		t.Fatalf("could not find both groups in:\n%s", strings.Join(lines, "\n"))
	}
	n := 0
	for _, l := range lines[first+1 : second] {
		if strings.TrimSpace(strings.ReplaceAll(l, zeroWidthSpace, "")) == "" {
			n++
		}
	}
	return n
}

// renderToPlain renders Markdown and strips the colour codes.
func renderToPlain(t *testing.T, md string) string {
	t.Helper()
	r, err := newRenderer("dark", 70)
	if err != nil {
		t.Fatal(err)
	}
	out, err := r.Render(md)
	if err != nil {
		t.Fatal(err)
	}
	return ansiCodes.ReplaceAllString(out, "")
}

// renderWithMarkers renders through the same marker-tagged style the preview
// uses, so link text can be located afterwards.
func renderWithMarkers(t *testing.T, md string) string {
	t.Helper()
	r, err := newRenderer("dark", 70)
	if err != nil {
		t.Fatal(err)
	}
	out, err := r.Render(md)
	if err != nil {
		t.Fatal(err)
	}
	return out
}
