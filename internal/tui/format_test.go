package tui

import "testing"

// Each formatting helper works on the line under the cursor and toggles back
// off when applied twice, matching the web toolbar's buttons.
func TestLineFormatting(t *testing.T) {
	for _, tc := range []struct {
		name  string
		fn    func(string, int) (string, int)
		in    string
		at    int
		want  string
		again string // applying it a second time
	}{
		{"heading", heading, "a line", 2, "# a line", "## a line"},
		{"bullet", bullet, "a line", 2, "- a line", "a line"},
		{"numbered", numbered, "a line", 2, "1. a line", "a line"},
		{"task", task, "a line", 2, "- [ ] a line", "- a line"},
		{"quote", quote, "a line", 2, "> a line", "a line"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := tc.fn(tc.in, tc.at)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
			if again, _ := tc.fn(got, tc.at+2); again != tc.again {
				t.Errorf("applying twice gave %q, want %q", again, tc.again)
			}
		})
	}
}

func TestHeadingCyclesAndWrapsBackToPlain(t *testing.T) {
	line := "a line"
	for i := 1; i <= 6; i++ {
		line, _ = heading(line, 0)
	}
	if line != "###### a line" {
		t.Fatalf("after six presses: %q", line)
	}
	if next, _ := heading(line, 0); next != "a line" {
		t.Errorf("a seventh press should clear it, got %q", next)
	}
}

// The list types convert between each other rather than stacking up.
func TestListTypesReplaceEachOther(t *testing.T) {
	v, _ := bullet("a line", 0)
	if v != "- a line" {
		t.Fatal(v)
	}
	if v, _ = numbered(v, 0); v != "1. a line" {
		t.Errorf("bullet -> numbered gave %q", v)
	}
	if v, _ = task(v, 0); v != "- [ ] a line" {
		t.Errorf("numbered -> task gave %q", v)
	}
	if v, _ = bullet(v, 0); v != "- a line" {
		t.Errorf("task -> bullet gave %q", v)
	}
}

func TestToggleTick(t *testing.T) {
	v := "- [ ] feed the cat"
	v, _ = toggleTick(v, 8)
	if v != "- [x] feed the cat" {
		t.Fatalf("ticking gave %q", v)
	}
	v, _ = toggleTick(v, 8)
	if v != "- [ ] feed the cat" {
		t.Errorf("unticking gave %q", v)
	}
	// A line that is not a task is left alone.
	if got, _ := toggleTick("just text", 3); got != "just text" {
		t.Errorf("a plain line was changed: %q", got)
	}
}

func TestIndentationIsKept(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"   nested", "   - nested"},
		{"   - nested", "   nested"},
	} {
		if got, _ := bullet(tc.in, 5); got != tc.want {
			t.Errorf("bullet(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFormattingActsOnTheCursorsLineOnly(t *testing.T) {
	v := "first\nsecond\nthird"
	got, _ := bullet(v, 8) // somewhere in "second"
	if got != "first\n- second\nthird" {
		t.Errorf("got %q", got)
	}
}

func TestCodeBlockWrapsTheLine(t *testing.T) {
	got, pos := codeBlock("fmt.Println()", 4)
	if got != "```\nfmt.Println()\n```" {
		t.Fatalf("got %q", got)
	}
	if pos != len("```\nfmt.Println()") {
		t.Errorf("cursor at %d", pos)
	}
	// On a blank line it opens an empty fence with the cursor inside.
	empty, epos := codeBlock("", 0)
	if empty != "```\n\n```" || epos != 4 {
		t.Errorf("empty fence = %q at %d", empty, epos)
	}
}

func TestRuleGoesBelowTheCurrentLine(t *testing.T) {
	got, _ := rule("some text", 4)
	if got != "some text\n\n---\n" {
		t.Errorf("got %q", got)
	}
}

func TestStrikethroughAndInlineCodeWrapWords(t *testing.T) {
	if got, _ := mark("cancel this", 8, "~~", "~~"); got != "cancel ~~this~~" {
		t.Errorf("strikethrough gave %q", got)
	}
	if got, _ := mark("the value", 6, "`", "`"); got != "the `value`" {
		t.Errorf("inline code gave %q", got)
	}
}
