package tui

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/prathmesh/notes/internal/store"
)

func newTestModel(t *testing.T) (model, *store.Store) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return New(st), st
}

// key builds the message Bubble Tea delivers for a single character.
func key(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func press(m model, msg tea.Msg) model {
	next, _ := m.Update(msg)
	return next.(model)
}

func TestListModeKeysSwitchMode(t *testing.T) {
	m, st := newTestModel(t)
	if _, err := st.Create("First", "one"); err != nil {
		t.Fatal(err)
	}
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, reloadedMsg{notes: mustList(t, st)})

	for _, tc := range []struct {
		name string
		key  rune
		want mode
	}{
		{"new note", 'n', modeEdit},
		{"help", '?', modeHelp},
		{"search", '/', modeSearch},
		{"delete confirm", 'd', modeConfirm},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := press(m, key(tc.key))
			if got.mode != tc.want {
				t.Errorf("pressing %q: mode = %v, want %v", tc.key, got.mode, tc.want)
			}
		})
	}
}

func TestCreateNoteFlow(t *testing.T) {
	m, st := newTestModel(t)
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})

	m = press(m, key('n'))
	if m.mode != modeEdit {
		t.Fatalf("expected edit mode, got %v", m.mode)
	}
	for _, r := range "hello world" {
		m = press(m, key(r))
	}
	if got := m.body.Value(); got != "hello world" {
		t.Fatalf("body = %q, want %q", got, "hello world")
	}

	m = press(m, tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.mode != modeList {
		t.Errorf("after save, mode = %v, want list", m.mode)
	}

	notes := mustList(t, st)
	if len(notes) != 1 {
		t.Fatalf("expected 1 note saved, got %d", len(notes))
	}
	if notes[0].Content != "hello world" {
		t.Errorf("content = %q", notes[0].Content)
	}
	if notes[0].Title != "hello world" {
		t.Errorf("title should fall back to the first line, got %q", notes[0].Title)
	}
}

func TestEscapeDiscardsEdit(t *testing.T) {
	m, st := newTestModel(t)
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, key('n'))
	m = press(m, key('x'))
	m = press(m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.mode != modeList {
		t.Errorf("mode = %v, want list", m.mode)
	}
	if notes := mustList(t, st); len(notes) != 0 {
		t.Errorf("escape should not save, but found %d notes", len(notes))
	}
}

func TestDeleteConfirmation(t *testing.T) {
	m, st := newTestModel(t)
	if _, err := st.Create("Doomed", "x"); err != nil {
		t.Fatal(err)
	}
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, reloadedMsg{notes: mustList(t, st)})

	// "n" at the confirmation prompt must cancel, not delete.
	cancelled := press(press(m, key('d')), key('n'))
	if cancelled.mode != modeList {
		t.Errorf("cancel: mode = %v, want list", cancelled.mode)
	}
	if len(mustList(t, st)) != 1 {
		t.Fatal("note was deleted despite cancelling")
	}

	confirmed := press(press(m, key('d')), key('y'))
	if confirmed.mode != modeList {
		t.Errorf("confirm: mode = %v, want list", confirmed.mode)
	}
	if n := len(mustList(t, st)); n != 0 {
		t.Errorf("note should be gone, %d remain", n)
	}
}

func TestNavigationClampsToBounds(t *testing.T) {
	m, st := newTestModel(t)
	for _, title := range []string{"a", "b", "c"} {
		if _, err := st.Create(title, title); err != nil {
			t.Fatal(err)
		}
	}
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, reloadedMsg{notes: mustList(t, st)})

	up := press(m, key('k')) // already at the top
	if up.cursor != 0 {
		t.Errorf("cursor went above the first note: %d", up.cursor)
	}
	down := m
	for i := 0; i < 10; i++ {
		down = press(down, key('j'))
	}
	if down.cursor != 2 {
		t.Errorf("cursor = %d, want it clamped to 2", down.cursor)
	}
}

func TestViewRendersWithoutWindowSize(t *testing.T) {
	m, _ := newTestModel(t)
	if out := m.View(); out == "" {
		t.Error("View() returned nothing before any WindowSizeMsg")
	}
}

func mustList(t *testing.T, st *store.Store) []store.Note {
	t.Helper()
	notes, err := st.List("")
	if err != nil {
		t.Fatalf("listing notes: %v", err)
	}
	return notes
}

func TestUndoRestoresLastDelete(t *testing.T) {
	m, st := newTestModel(t)
	n, err := st.Create("Precious", "body")
	if err != nil {
		t.Fatal(err)
	}
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, reloadedMsg{notes: mustList(t, st)})

	m = press(press(m, key('d')), key('y'))
	if len(mustList(t, st)) != 0 {
		t.Fatal("note was not trashed")
	}

	m = press(m, key('u'))
	notes := mustList(t, st)
	if len(notes) != 1 || notes[0].ID != n.ID {
		t.Fatalf("undo did not restore the note: %v", notes)
	}
}

func TestUndoWithNothingDeletedIsHarmless(t *testing.T) {
	m, st := newTestModel(t)
	if _, err := st.Create("a", "b"); err != nil {
		t.Fatal(err)
	}
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, reloadedMsg{notes: mustList(t, st)})

	m = press(m, key('u'))
	if m.err != nil {
		t.Errorf("undo with nothing to undo set an error: %v", m.err)
	}
	if len(mustList(t, st)) != 1 {
		t.Error("undo changed the notes")
	}
}
