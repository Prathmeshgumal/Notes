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
