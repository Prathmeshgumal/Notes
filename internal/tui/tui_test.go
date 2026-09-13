package tui

import (
	"path/filepath"
	"strings"
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

func TestListIsMostRecentlyEditedFirst(t *testing.T) {
	m, st := newTestModel(t)
	first, _ := st.Create("oldest", "a")
	if _, err := st.Create("newest", "b"); err != nil {
		t.Fatal(err)
	}
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, reloadedMsg{notes: mustList(t, st)})
	if m.notes[0].Title != "newest" {
		t.Fatalf("order = %v, want the newest note first", titlesOf(m.notes))
	}

	// Touching the older note must float it to the top.
	if _, err := st.Update(first.ID, "oldest", "edited"); err != nil {
		t.Fatal(err)
	}
	m = press(m, reloadedMsg{notes: mustList(t, st)})
	if m.notes[0].Title != "oldest" {
		t.Errorf("order = %v, want the just-edited note first", titlesOf(m.notes))
	}
}

// Saving reorders the list; the cursor must follow the note, not the index.
func TestSelectionFollowsNoteAcrossReorder(t *testing.T) {
	m, st := newTestModel(t)
	if _, err := st.Create("one", "a"); err != nil {
		t.Fatal(err)
	}
	older, _ := st.Create("two", "b")
	if _, err := st.Create("three", "c"); err != nil {
		t.Fatal(err)
	}

	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, reloadedMsg{notes: mustList(t, st)})

	// Select the oldest note, which sits last.
	for m.selected() != nil && m.selected().ID != older.ID {
		m = press(m, key('j'))
	}
	if m.selected() == nil || m.selected().ID != older.ID {
		t.Fatal("could not select the target note")
	}

	// Edit and save it: it jumps to the top of the list.
	m = press(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = press(m, key('!'))
	m = press(m, tea.KeyMsg{Type: tea.KeyCtrlS})
	m = press(m, reloadedMsg{notes: mustList(t, st)})

	if got := m.selected(); got == nil || got.ID != older.ID {
		t.Errorf("selection landed on %v, want the note that was just saved", got)
	}
}

func titlesOf(notes []store.Note) []string {
	out := make([]string, len(notes))
	for i, n := range notes {
		out[i] = n.Title
	}
	return out
}
func TestCtrlBBoldsTheWordUnderTheCursor(t *testing.T) {
	m, _ := newTestModel(t)
	m = press(m, tea.WindowSizeMsg{Width: 90, Height: 30})
	m = press(m, key('n'))
	for _, r := range "hello" {
		m = press(m, key(r))
	}
	m = press(m, tea.KeyMsg{Type: tea.KeyCtrlB})

	if got := m.body.Value(); got != "**hello**" {
		t.Errorf("body = %q, want %q", got, "**hello**")
	}
}

func TestAltIItalicsTheWord(t *testing.T) {
	m, _ := newTestModel(t)
	m = press(m, tea.WindowSizeMsg{Width: 90, Height: 30})
	m = press(m, key('n'))
	for _, r := range "hello" {
		m = press(m, key(r))
	}
	m = press(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}, Alt: true})

	if got := m.body.Value(); got != "*hello*" {
		t.Errorf("body = %q, want %q", got, "*hello*")
	}
}

func TestCtrlKMakesALink(t *testing.T) {
	m, _ := newTestModel(t)
	m = press(m, tea.WindowSizeMsg{Width: 90, Height: 30})
	m = press(m, key('n'))
	for _, r := range "docs" {
		m = press(m, key(r))
	}
	m = press(m, tea.KeyMsg{Type: tea.KeyCtrlK})

	if got := m.body.Value(); got != "[docs]()" {
		t.Errorf("body = %q, want %q", got, "[docs]()")
	}
	// Typing continues inside the parentheses.
	for _, r := range "https://x.test" {
		m = press(m, key(r))
	}
	if got := m.body.Value(); got != "[docs](https://x.test)" {
		t.Errorf("typing after ctrl+k gave %q", got)
	}
}

// The title field is a single line; formatting keys there would be noise.
func TestFormattingKeysAreIgnoredInTheTitle(t *testing.T) {
	m, _ := newTestModel(t)
	m = press(m, tea.WindowSizeMsg{Width: 90, Height: 30})
	m = press(m, key('n'))
	m = press(m, tea.KeyMsg{Type: tea.KeyTab}) // focus the title
	for _, r := range "plain" {
		m = press(m, key(r))
	}
	m = press(m, tea.KeyMsg{Type: tea.KeyCtrlB})

	if got := m.title.Value(); got != "plain" {
		t.Errorf("title = %q, want it untouched", got)
	}
}
func stripANSI(s string) string {
	var b strings.Builder
	esc := false
	for _, r := range s {
		if r == 0x1b {
			esc = true
			continue
		}
		if esc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				esc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Deleting several notes must be undoable several times, not just once.
func TestUndoWalksBackThroughDeletes(t *testing.T) {
	m, st := newTestModel(t)
	for _, title := range []string{"one", "two", "three"} {
		if _, err := st.Create(title, title); err != nil {
			t.Fatal(err)
		}
	}
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, reloadedMsg{notes: mustList(t, st)})

	for i := 0; i < 3; i++ {
		m = press(press(m, key('d')), key('y'))
		m = press(m, reloadedMsg{notes: mustList(t, st)})
	}
	if n := len(mustList(t, st)); n != 0 {
		t.Fatalf("expected everything trashed, %d left", n)
	}

	for i := 1; i <= 3; i++ {
		m = press(m, key('u'))
		m = press(m, reloadedMsg{notes: mustList(t, st)})
		if got := len(mustList(t, st)); got != i {
			t.Fatalf("after %d undos, %d notes are back, want %d", i, got, i)
		}
	}

	// A fourth undo has nothing left to do and must not error.
	m = press(m, key('u'))
	if m.err != nil {
		t.Errorf("undo past the end set an error: %v", m.err)
	}
}

func TestTrashViewListsAndRestores(t *testing.T) {
	m, st := newTestModel(t)
	keep, _ := st.Create("keep", "a")
	gone, _ := st.Create("gone", "b")
	if err := st.Delete(gone.ID); err != nil {
		t.Fatal(err)
	}

	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, reloadedMsg{notes: mustList(t, st)})

	m = press(m, key('T'))
	if m.mode != modeTrash {
		t.Fatalf("T should open the trash, mode = %v", m.mode)
	}
	trash, err := st.Trash()
	if err != nil {
		t.Fatal(err)
	}
	m = press(m, trashLoadedMsg{notes: trash})
	if len(m.trash) != 1 || m.trash[0].ID != gone.ID {
		t.Fatalf("trash holds %v, want the deleted note", titlesOf(m.trash))
	}

	m = press(m, tea.KeyMsg{Type: tea.KeyEnter})
	back := mustList(t, st)
	if len(back) != 2 {
		t.Fatalf("after restoring, %d notes are live, want 2", len(back))
	}
	_ = keep

	m = press(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.mode != modeList {
		t.Errorf("esc should leave the trash, mode = %v", m.mode)
	}
}

// Restoring from the trash must not leave a stale id for undo to trip over.
func TestUndoAfterRestoringFromTrashIsHarmless(t *testing.T) {
	m, st := newTestModel(t)
	n, _ := st.Create("note", "x")
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, reloadedMsg{notes: mustList(t, st)})

	m = press(press(m, key('d')), key('y'))  // trashed, id pushed for undo
	if err := st.Restore(n.ID); err != nil { // restored another way
		t.Fatal(err)
	}

	m = press(m, key('u'))
	if m.err != nil {
		t.Errorf("undo of an already-restored note errored: %v", m.err)
	}
	if got := len(mustList(t, st)); got != 1 {
		t.Errorf("%d notes live, want 1", got)
	}
}

func TestPurgeFromTrashNeedsConfirmation(t *testing.T) {
	m, st := newTestModel(t)
	n, _ := st.Create("doomed", "x")
	if err := st.Delete(n.ID); err != nil {
		t.Fatal(err)
	}
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, key('T'))
	trash, _ := st.Trash()
	m = press(m, trashLoadedMsg{notes: trash})

	// Answering no must leave it recoverable.
	cancelled := press(press(m, key('d')), key('n'))
	if cancelled.mode != modeTrash {
		t.Errorf("cancelling should return to the trash, mode = %v", cancelled.mode)
	}
	if tr, _ := st.Trash(); len(tr) != 1 {
		t.Fatal("the note was purged despite cancelling")
	}

	confirmed := press(press(m, key('d')), key('y'))
	if confirmed.mode != modeTrash {
		t.Errorf("confirming should return to the trash, mode = %v", confirmed.mode)
	}
	if tr, _ := st.Trash(); len(tr) != 0 {
		t.Errorf("the note was not purged, %d left in the trash", len(tr))
	}
	// Gone for good.
	if err := st.Restore(n.ID); err == nil {
		t.Error("a purged note could still be restored")
	}
}

func TestEmptyTrashNeedsConfirmationAndSparesLiveNotes(t *testing.T) {
	m, st := newTestModel(t)
	if _, err := st.Create("keep me", "a"); err != nil {
		t.Fatal(err)
	}
	for _, title := range []string{"x", "y"} {
		n, _ := st.Create(title, title)
		if err := st.Delete(n.ID); err != nil {
			t.Fatal(err)
		}
	}
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, key('T'))
	trash, _ := st.Trash()
	m = press(m, trashLoadedMsg{notes: trash})

	cancelled := press(press(m, key('E')), key('n'))
	if tr, _ := st.Trash(); len(tr) != 2 {
		t.Fatal("the trash was emptied despite cancelling")
	}
	_ = cancelled

	m = press(press(m, key('E')), key('y'))
	if tr, _ := st.Trash(); len(tr) != 0 {
		t.Errorf("trash still holds %d notes", len(tr))
	}
	live := mustList(t, st)
	if len(live) != 1 || live[0].Title != "keep me" {
		t.Errorf("live notes = %v, want only the kept one", titlesOf(live))
	}
	// Undo must not try to restore ids that no longer exist.
	m = press(m, tea.KeyMsg{Type: tea.KeyEsc})
	m = press(m, key('u'))
	if m.err != nil {
		t.Errorf("undo after emptying the trash errored: %v", m.err)
	}
}

// The question must name what it is about, so nothing is confirmed blind.
func TestConfirmPromptsNameTheirTarget(t *testing.T) {
	m, st := newTestModel(t)
	n, _ := st.Create("Quarterly plan", "x")
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, reloadedMsg{notes: mustList(t, st)})

	asked := press(m, key('d'))
	if !strings.Contains(asked.confirmPrompt, "Quarterly plan") {
		t.Errorf("prompt %q does not name the note", asked.confirmPrompt)
	}
	if !strings.Contains(asked.View(), "Quarterly plan") {
		t.Error("the question is not visible on screen")
	}

	if err := st.Delete(n.ID); err != nil {
		t.Fatal(err)
	}
	inTrash := press(m, key('T'))
	trash, _ := st.Trash()
	inTrash = press(inTrash, trashLoadedMsg{notes: trash})
	purge := press(inTrash, key('d'))
	if !strings.Contains(purge.confirmPrompt, "cannot be undone") {
		t.Errorf("a permanent delete should say so: %q", purge.confirmPrompt)
	}
}
