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

// R shows the Markdown source, so it can be selected with the mouse and
// pasted elsewhere even where no clipboard route works.
func TestRawViewShowsTheSource(t *testing.T) {
	m, st := newTestModel(t)
	if _, err := st.Create("Note", "# Heading\n\n**bold** text"); err != nil {
		t.Fatal(err)
	}
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 24})
	m = press(m, reloadedMsg{notes: mustList(t, st)})

	rendered := stripANSI(m.View())
	if strings.Contains(rendered, "**bold**") {
		t.Error("the rendered view is showing raw Markdown")
	}

	m = press(m, key('R'))
	raw := stripANSI(m.View())
	if !strings.Contains(raw, "**bold**") {
		t.Errorf("R did not show the source:\n%s", raw)
	}

	m = press(m, key('R'))
	if strings.Contains(stripANSI(m.View()), "**bold**") {
		t.Error("R did not switch back to the rendered view")
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
