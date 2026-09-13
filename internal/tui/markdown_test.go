package tui

import (
	"strings"
	"testing"
)

func TestSeparateListGroups(t *testing.T) {
	for _, tc := range []struct {
		name   string
		in     string
		expect bool // should a separator be inserted?
	}{
		{
			name:   "blank line between two bullet groups",
			in:     "- one\n- two\n\n- three\n",
			expect: true,
		},
		{
			name:   "task lists with nested children",
			in:     "- [ ] parent\n   - [ ] child\n\n- [ ] other\n   - [ ] child\n",
			expect: true,
		},
		{
			name:   "ordered lists are left alone",
			in:     "1. one\n2. two\n\n3. three\n",
			expect: false,
		},
		{
			name:   "prose after a list is not a list group",
			in:     "- one\n- two\n\nJust a paragraph.\n",
			expect: false,
		},
		{
			name:   "a list following a paragraph is untouched",
			in:     "Some intro.\n\n- one\n- two\n",
			expect: false,
		},
		{
			name:   "no blank lines at all",
			in:     "- one\n- two\n- three\n",
			expect: false,
		},
		{
			name:   "blank line inside a fenced code block",
			in:     "```\n- one\n\n- two\n```\n",
			expect: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := separateListGroups(tc.in)
			has := strings.Contains(got, listSeparator)
			if has != tc.expect {
				t.Errorf("separator inserted = %v, want %v\ninput:\n%s\ngot:\n%s",
					has, tc.expect, tc.in, got)
			}
		})
	}
}

// The note itself must never change — only what we hand to the renderer.
func TestSeparateListGroupsPreservesContent(t *testing.T) {
	in := "- [ ] one\n   - [ ] nested\n\n- [ ] two\n"
	got := separateListGroups(in)

	// Remove the separator and the blank line it needs around it, then the
	// text handed to the renderer must match the note exactly.
	stripped := strings.ReplaceAll(got, listSeparator+"\n", "")
	for strings.Contains(stripped, "\n\n\n") {
		stripped = strings.ReplaceAll(stripped, "\n\n\n", "\n\n")
	}
	if strings.TrimSpace(stripped) != strings.TrimSpace(in) {
		t.Errorf("content changed beyond the separator:\nwant %q\ngot  %q", in, stripped)
	}
}

func TestSeparateListGroupsHandlesMultipleGroups(t *testing.T) {
	in := "- a\n\n- b\n\n- c\n"
	got := separateListGroups(in)
	if n := strings.Count(got, listSeparator); n != 2 {
		t.Errorf("got %d separators, want 2:\n%s", n, got)
	}
}

// The point of the fillers is vertical space, so assert on what glamour
// actually renders rather than on the intermediate Markdown.
func TestBlankRunHeightIsPreserved(t *testing.T) {
	for _, tc := range []struct{ typed, rendered int }{
		{typed: 1, rendered: 1},
		{typed: 2, rendered: 1},
		{typed: 3, rendered: 3},
		{typed: 5, rendered: 5},
		{typed: 7, rendered: 7},
	} {
		md := "- [ ] first group\n" + strings.Repeat("\n", tc.typed) + "- [ ] second group\n"
		got := renderedBlankLines(t, separateListGroups(md))
		if got != tc.rendered {
			t.Errorf("typing %d blank lines rendered %d, want %d", tc.typed, got, tc.rendered)
		}
	}
}

func TestFillersFor(t *testing.T) {
	for _, tc := range []struct{ run, want int }{
		{0, 0}, {1, 0}, {2, 0}, {3, 1}, {4, 1}, {5, 2}, {7, 3},
	} {
		if got := fillersFor(tc.run); got != tc.want {
			t.Errorf("fillersFor(%d) = %d, want %d", tc.run, got, tc.want)
		}
	}
}

func TestStripDerivedTitle(t *testing.T) {
	for _, tc := range []struct {
		name    string
		content string
		title   string
		want    string
	}{
		{
			name:    "plain first line matching the title is dropped",
			content: "14 sep Tasks\n\n- [ ] one\n",
			title:   "14 sep Tasks",
			want:    "- [ ] one\n",
		},
		{
			name:    "heading matching the title is dropped",
			content: "# Welcome\n\nbody text\n",
			title:   "Welcome",
			want:    "body text\n",
		},
		{
			name:    "an explicit different title leaves the body alone",
			content: "# Welcome\n\nbody text\n",
			title:   "Something else",
			want:    "# Welcome\n\nbody text\n",
		},
		{
			name:    "a body that does not start with the title is untouched",
			content: "- [ ] one\n- [ ] two\n",
			title:   "My list",
			want:    "- [ ] one\n- [ ] two\n",
		},
		{
			name:    "leading blank lines are handled",
			content: "\n\nNotes\n\ncontent\n",
			title:   "Notes",
			want:    "content\n",
		},
		{
			name:    "a single-line note becomes empty",
			content: "just this\n",
			title:   "just this",
			want:    "",
		},
		{
			name:    "empty title changes nothing",
			content: "anything\n",
			title:   "",
			want:    "anything\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripDerivedTitle(tc.content, tc.title); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// A single newline breaks the line, matching this app's web UI and Gists.
// Without this the renderer runs consecutive lines together into a paragraph.
func TestSingleNewlineBreaksTheLine(t *testing.T) {
	out := renderToPlain(t, "first line\nsecond line\n")
	lines := []string{}
	for _, l := range strings.Split(out, "\n") {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, strings.TrimSpace(l))
		}
	}
	if len(lines) != 2 {
		t.Fatalf("expected two lines, got %d:\n%q", len(lines), out)
	}
	if lines[0] != "first line" || lines[1] != "second line" {
		t.Errorf("lines were not kept apart: %q", lines)
	}
}

func TestTableNeedsItsSeparatorRow(t *testing.T) {
	withSep := renderToPlain(t, "|col1|col2|\n|----|----|\n|10  |20  |\n")
	if !strings.Contains(withSep, "│") {
		t.Errorf("a table with a separator row should render as a table:\n%s", withSep)
	}

	// Without the separator row it is not a table in GitHub-flavoured Markdown,
	// and stays plain text. Asserted so the behaviour is documented, not lost.
	without := renderToPlain(t, "|col1|col2|\n|10  |20  |\n")
	if strings.Contains(without, "│") {
		t.Errorf("without a separator row this should not become a table:\n%s", without)
	}
}

// A task line carries a bullet as well as its checkbox. This departs from
// GitHub, which hides the bullet, and is deliberate.
func TestTaskLinesShowBulletThenCheckbox(t *testing.T) {
	out := renderToPlain(t, "- [x] done\n- [ ] open\n- plain bullet\n")

	for _, want := range []string{"• ☑ done", "• ☐ open", "• plain bullet"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestCheckboxesUseBoxGlyphs(t *testing.T) {
	out := renderToPlain(t, "- [x] done\n- [ ] open\n")
	if !strings.Contains(out, "☑") {
		t.Errorf("a completed task should show a ticked box:\n%s", out)
	}
	if !strings.Contains(out, "☐") {
		t.Errorf("an open task should show an empty box:\n%s", out)
	}
	if strings.Contains(out, "[ ]") || strings.Contains(out, "[✓]") {
		t.Errorf("the bracketed markers should be gone:\n%s", out)
	}
}
