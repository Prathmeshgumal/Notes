package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteIsRecoverable(t *testing.T) {
	st := newTestStore(t)
	n, err := st.Create("Important", "do not lose me")
	if err != nil {
		t.Fatal(err)
	}

	if err := st.Delete(n.ID); err != nil {
		t.Fatal(err)
	}
	if got := len(mustListStore(t, st)); got != 0 {
		t.Errorf("deleted note still listed (%d notes)", got)
	}
	if _, err := st.Get(n.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get on a trashed note = %v, want ErrNotFound", err)
	}

	trash, err := st.Trash()
	if err != nil {
		t.Fatal(err)
	}
	if len(trash) != 1 || trash[0].ID != n.ID {
		t.Fatalf("trash = %v, want the deleted note", trash)
	}

	if err := st.Restore(n.ID); err != nil {
		t.Fatal(err)
	}
	back, err := st.Get(n.ID)
	if err != nil {
		t.Fatalf("restored note not readable: %v", err)
	}
	if back.Content != "do not lose me" {
		t.Errorf("content changed through the trash: %q", back.Content)
	}
}

func TestTrashedNotesAreExcludedFromSearch(t *testing.T) {
	st := newTestStore(t)
	n, _ := st.Create("Secret", "hidden text")
	if err := st.Delete(n.ID); err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{"secret", "hidden"} {
		if got, _ := st.List(q); len(got) != 0 {
			t.Errorf("List(%q) returned a trashed note", q)
		}
	}
	if c, _ := st.Count(); c != 0 {
		t.Errorf("Count() = %d, want trashed notes excluded", c)
	}
}

func TestRestoreUnknownIDFails(t *testing.T) {
	st := newTestStore(t)
	if err := st.Restore("nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Restore = %v, want ErrNotFound", err)
	}
	n, _ := st.Create("live", "x")
	if err := st.Restore(n.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("restoring a note that is not trashed = %v, want ErrNotFound", err)
	}
}

func TestSnapshotCopiesAndPrunes(t *testing.T) {
	dir := t.TempDir()
	db := filepath.Join(dir, "notes.db")

	st, err := Open(db)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.Create("keep me", "body"); err != nil {
		t.Fatal(err)
	}
	st.Close()

	// More snapshots than the retention limit; names carry a 1s timestamp, so
	// write them directly to control the count deterministically.
	backups := filepath.Join(dir, "backups")
	if _, err := Snapshot(db); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(backups)
	if err != nil {
		t.Fatalf("no backups directory: %v", err)
	}
	if got := countSnapshots(t, backups); got != 1 {
		t.Fatalf("expected 1 snapshot, got %d", got)
	}

	// A snapshot must be a readable database holding the same note.
	snap := filepath.Join(backups, entries[0].Name())
	restored, err := Open(snap)
	if err != nil {
		t.Fatalf("snapshot is not a usable database: %v", err)
	}
	defer restored.Close()
	notes := mustListStore(t, restored)
	if len(notes) != 1 || notes[0].Title != "keep me" {
		t.Errorf("snapshot contents = %v, want the original note", notes)
	}

	for i := 0; i < SnapshotsKept+3; i++ {
		name := filepath.Join(backups, fmt.Sprintf("notes-20200101-%06d.db", i))
		if err := copyFile(db, name); err != nil {
			t.Fatal(err)
		}
	}
	before := countSnapshots(t, backups)
	if err := prune(backups); err != nil {
		t.Fatal(err)
	}
	// Only .db files are snapshots; opening one leaves -wal/-shm beside it.
	if after := countSnapshots(t, backups); after > SnapshotsKept {
		t.Errorf("prune left %d of %d snapshots, want at most %d",
			after, before, SnapshotsKept)
	}
}

func TestSnapshotOfMissingFileIsNotAnError(t *testing.T) {
	if _, err := Snapshot(filepath.Join(t.TempDir(), "absent.db")); err != nil {
		t.Errorf("Snapshot on a missing database = %v, want nil", err)
	}
}

func countSnapshots(t *testing.T, dir string) int {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".db" {
			n++
		}
	}
	return n
}

func mustListStore(t *testing.T, st *Store) []Note {
	t.Helper()
	notes, err := st.List("")
	if err != nil {
		t.Fatal(err)
	}
	return notes
}

func TestPurgeRemovesOnlyTrashedNotes(t *testing.T) {
	st := newTestStore(t)
	live, _ := st.Create("live", "a")
	gone, _ := st.Create("gone", "b")
	if err := st.Delete(gone.ID); err != nil {
		t.Fatal(err)
	}

	// A live note must not be purgeable.
	if err := st.Purge(live.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("purging a live note = %v, want ErrNotFound", err)
	}
	if len(mustListStore(t, st)) != 1 {
		t.Fatal("the live note was removed")
	}

	if err := st.Purge(gone.ID); err != nil {
		t.Fatal(err)
	}
	trash, _ := st.Trash()
	if len(trash) != 0 {
		t.Errorf("trash still holds %d notes", len(trash))
	}
	// Gone for good: it cannot be restored.
	if err := st.Restore(gone.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("a purged note came back: %v", err)
	}
}

func TestEmptyTrashLeavesLiveNotesAlone(t *testing.T) {
	st := newTestStore(t)
	if _, err := st.Create("keep me", "a"); err != nil {
		t.Fatal(err)
	}
	for _, title := range []string{"x", "y", "z"} {
		n, _ := st.Create(title, title)
		if err := st.Delete(n.ID); err != nil {
			t.Fatal(err)
		}
	}

	n, err := st.EmptyTrash()
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("emptied %d notes, want 3", n)
	}
	if trash, _ := st.Trash(); len(trash) != 0 {
		t.Errorf("trash is not empty: %d", len(trash))
	}
	live := mustListStore(t, st)
	if len(live) != 1 || live[0].Title != "keep me" {
		t.Errorf("live notes = %v, want just the kept one", titles(live))
	}
}

func TestEmptyTrashOnAnEmptyTrash(t *testing.T) {
	st := newTestStore(t)
	n, err := st.EmptyTrash()
	if err != nil || n != 0 {
		t.Errorf("EmptyTrash on an empty trash = (%d, %v), want (0, nil)", n, err)
	}
}
