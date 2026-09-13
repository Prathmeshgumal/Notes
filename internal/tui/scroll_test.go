package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func longNoteModel(t *testing.T) model {
	t.Helper()
	m, st := newTestModel(t)
	body := "# Long note\n\n" + strings.Repeat("a line of text\n", 200)
	if _, err := st.Create("Long note", body); err != nil {
		t.Fatal(err)
	}
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 24})
	return press(m, reloadedMsg{notes: mustList(t, st)})
}

// A terminal turns the mouse wheel into arrow keys, so the arrows must scroll
// the note rather than jump between notes.
func TestArrowsScrollThePreview(t *testing.T) {
	m := longNoteModel(t)
	if m.preview.TotalLineCount() <= m.preview.Height {
		t.Fatal("the test note is not longer than the pane")
	}

	start := m.cursor
	for i := 0; i < 12; i++ {
		m = press(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.preview.YOffset != 12 {
		t.Errorf("12 down presses scrolled %d lines, want 12", m.preview.YOffset)
	}
	if m.cursor != start {
		t.Errorf("scrolling changed the selected note (%d -> %d)", start, m.cursor)
	}

	for i := 0; i < 5; i++ {
		m = press(m, tea.KeyMsg{Type: tea.KeyUp})
	}
	if m.preview.YOffset != 7 {
		t.Errorf("after scrolling back, offset = %d, want 7", m.preview.YOffset)
	}
}

// j and k still move between notes, and do not scroll.
func TestJKMoveBetweenNotes(t *testing.T) {
	m, st := newTestModel(t)
	for _, title := range []string{"one", "two", "three"} {
		if _, err := st.Create(title, strings.Repeat(title+"\n", 100)); err != nil {
			t.Fatal(err)
		}
	}
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 24})
	m = press(m, reloadedMsg{notes: mustList(t, st)})

	m = press(m, key('j'))
	if m.cursor != 1 {
		t.Errorf("j moved to %d, want 1", m.cursor)
	}
	if m.preview.YOffset != 0 {
		t.Errorf("moving to another note left the preview scrolled to %d", m.preview.YOffset)
	}
	m = press(m, key('k'))
	if m.cursor != 0 {
		t.Errorf("k moved to %d, want 0", m.cursor)
	}
}

func TestPageAndHomeEndScrolling(t *testing.T) {
	m := longNoteModel(t)
	total := m.preview.TotalLineCount()

	m = press(m, tea.KeyMsg{Type: tea.KeyPgDown})
	if m.preview.YOffset == 0 {
		t.Error("page down did not scroll")
	}
	m = press(m, tea.KeyMsg{Type: tea.KeyEnd})
	if m.preview.YOffset != total-m.preview.Height {
		t.Errorf("end left offset at %d, want the bottom (%d)",
			m.preview.YOffset, total-m.preview.Height)
	}
	m = press(m, tea.KeyMsg{Type: tea.KeyHome})
	if m.preview.YOffset != 0 {
		t.Errorf("home left offset at %d, want 0", m.preview.YOffset)
	}
}

// R opens a bare, full-width view of the Markdown source. It carries no
// borders or padding, so selecting it with the mouse copies the note and
// nothing else — which is the only route that works when no clipboard tool is
// installed and the terminal refuses the clipboard escape.
func TestRawViewIsBareSource(t *testing.T) {
	m, st := newTestModel(t)
	src := "# Heading\n\n**bold** text\n- [ ] a task"
	if _, err := st.Create("Note", src); err != nil {
		t.Fatal(err)
	}
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 24})
	m = press(m, reloadedMsg{notes: mustList(t, st)})

	if strings.Contains(stripANSI(m.View()), "**bold**") {
		t.Error("the rendered view is showing raw Markdown")
	}

	m = press(m, key('R'))
	if m.mode != modeRaw {
		t.Fatalf("R did not open the source view, mode = %v", m.mode)
	}
	view := stripANSI(m.View())

	for _, want := range []string{"# Heading", "**bold** text", "- [ ] a task"} {
		if !strings.Contains(view, want) {
			t.Errorf("source view is missing %q:\n%s", want, view)
		}
	}
	// Nothing to catch in a selection but the note itself.
	for _, chrome := range []string{"│", "╭", "╰", "┃"} {
		if strings.Contains(view, chrome) {
			t.Errorf("the source view draws %q, which a mouse selection would copy:\n%s",
				chrome, view)
		}
	}

	m = press(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.mode != modeList {
		t.Error("esc did not leave the source view")
	}
}

func TestRawViewScrolls(t *testing.T) {
	m, st := newTestModel(t)
	if _, err := st.Create("Long", strings.Repeat("a line\n", 300)); err != nil {
		t.Fatal(err)
	}
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 24})
	m = press(m, reloadedMsg{notes: mustList(t, st)})
	m = press(m, key('R'))

	for i := 0; i < 10; i++ {
		m = press(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.rawView.YOffset != 10 {
		t.Errorf("source view scrolled %d lines, want 10", m.rawView.YOffset)
	}
	m = press(m, tea.KeyMsg{Type: tea.KeyEnd})
	if m.rawView.YOffset == 10 {
		t.Error("end did not jump to the bottom")
	}
}

func TestScrollbar(t *testing.T) {
	// Everything fits: a blank column, not a full-height bar.
	for _, cell := range scrollbar(10, 5, 0) {
		if cell != "" {
			t.Errorf("a note that fits should get no bar, got %q", cell)
		}
	}

	thumbAt := func(bar []string) (first, count int) {
		first = -1
		for i, c := range bar {
			if strings.Contains(c, "┃") {
				if first < 0 {
					first = i
				}
				count++
			}
		}
		return
	}

	top, n := thumbAt(scrollbar(10, 100, 0))
	if top != 0 || n < 1 {
		t.Errorf("at the top the thumb should start at 0, got %d (size %d)", top, n)
	}

	bottom, n2 := thumbAt(scrollbar(10, 100, 90))
	if bottom+n2 != 10 {
		t.Errorf("at the bottom the thumb should reach the end, got %d+%d", bottom, n2)
	}

	mid, _ := thumbAt(scrollbar(10, 100, 45))
	if mid <= top || mid >= bottom {
		t.Errorf("halfway the thumb should sit between the ends: %d (top %d, bottom %d)",
			mid, top, bottom)
	}

	// A very long note still gets a visible thumb.
	if _, n := thumbAt(scrollbar(10, 100000, 0)); n < 1 {
		t.Error("the thumb vanished on a very long note")
	}
}

func TestScrollbarIsAppendedToEveryLine(t *testing.T) {
	content := strings.Join([]string{"one", "two", "three"}, "\n")
	out := withScrollbar(content, 3, 30, 0)
	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	for i, l := range lines {
		if !strings.Contains(l, "┃") && !strings.Contains(l, "│") {
			t.Errorf("line %d has no scrollbar cell: %q", i, l)
		}
	}
}

// The key list outgrew a screen, so it has to scroll — and a keystroke meant
// for scrolling must not close it.
func TestHelpScrolls(t *testing.T) {
	m, _ := newTestModel(t)
	m = press(m, tea.WindowSizeMsg{Width: 90, Height: 24})
	m = press(m, key('?'))
	if m.mode != modeHelp {
		t.Fatalf("? did not open the help, mode = %v", m.mode)
	}
	if m.help.TotalLineCount() <= m.help.Height {
		t.Fatal("the help now fits on screen; this test is no longer meaningful")
	}

	top := stripANSI(m.View())
	if !strings.Contains(top, "Choosing a note") {
		t.Errorf("the help does not start at the top:\n%s", top)
	}

	for i := 0; i < 10; i++ {
		m = press(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.mode != modeHelp {
		t.Fatal("scrolling closed the help")
	}
	if m.help.YOffset != 10 {
		t.Errorf("scrolled %d lines, want 10", m.help.YOffset)
	}

	m = press(m, tea.KeyMsg{Type: tea.KeyEnd})
	if !strings.Contains(stripANSI(m.View()), "quit") {
		t.Error("the end of the help is not reachable")
	}

	m = press(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.mode != modeList {
		t.Error("esc did not close the help")
	}
}

// Everything the app binds should be findable in the help.
func TestHelpMentionsTheListBehaviour(t *testing.T) {
	for _, want := range []string{
		"Lists carry on by themselves",
		"- [ ] buy milk",
		"keeps counting",
		"removes the marker",
		"alt+↵",
		"Shift+Enter",
	} {
		if !strings.Contains(helpText, want) {
			t.Errorf("the help does not mention %q", want)
		}
	}
}

// The list moved to a box in the top-right so the note gets the rest of the
// screen; the corner beneath it carries what is known about the note.
func TestAsideLayout(t *testing.T) {
	m, st := newTestModel(t)
	if _, err := st.Create("Tasks", "- [x] one\n- [x] two\n- [ ] three\nsome words here"); err != nil {
		t.Fatal(err)
	}
	m = press(m, tea.WindowSizeMsg{Width: 96, Height: 24})
	m = press(m, reloadedMsg{notes: mustList(t, st)})

	view := stripANSI(m.View())
	lines := strings.Split(view, "\n")

	// The note's own title should sit far to the left, the list far to the right.
	var titleCol, listCol int = -1, -1
	for _, l := range lines {
		if i := strings.Index(l, "Tasks"); i >= 0 && titleCol < 0 {
			titleCol = len([]rune(l[:i]))
		}
		if i := strings.Index(l, "Notes ("); i >= 0 && listCol < 0 {
			listCol = len([]rune(l[:i]))
		}
	}
	if titleCol < 0 || listCol < 0 {
		t.Fatalf("could not find both panes:\n%s", view)
	}
	if listCol <= titleCol {
		t.Errorf("the note list is at column %d, left of the note at %d", listCol, titleCol)
	}

	for _, want := range []string{"Edited", "Tasks", "2 of 3 done", "Length"} {
		if !strings.Contains(view, want) {
			t.Errorf("the details box is missing %q:\n%s", want, view)
		}
	}

	// The facts sit above the list, not below it.
	factsRow, listRow := -1, -1
	for i, l := range lines {
		if strings.Contains(l, "Edited") && factsRow < 0 {
			factsRow = i
		}
		if strings.Contains(l, "Notes (") && listRow < 0 {
			listRow = i
		}
	}
	if factsRow < 0 || listRow < 0 {
		t.Fatalf("could not find both right-hand boxes:\n%s", view)
	}
	if factsRow > listRow {
		t.Errorf("the facts box is below the list (rows %d vs %d)", factsRow, listRow)
	}
}

func TestCountTasks(t *testing.T) {
	for _, tc := range []struct {
		in          string
		done, total int
	}{
		{"- [x] a\n- [ ] b", 1, 2},
		{"- [X] a\n- [x] b", 2, 2},
		{"   - [ ] nested", 0, 1},
		{"no tasks here", 0, 0},
		{"- a plain bullet", 0, 0},
	} {
		d, n := countTasks(tc.in)
		if d != tc.done || n != tc.total {
			t.Errorf("countTasks(%q) = (%d,%d), want (%d,%d)", tc.in, d, n, tc.done, tc.total)
		}
	}
}
