package store

import (
	"fmt"
	"os"
	"path/filepath"
)

// AdoptLegacy moves a database left behind by an earlier name of this program
// into its new home, so an existing user's notes follow the rename without
// them having to do anything.
//
// It only ever acts when there is nothing at the new path: an existing database
// is never touched, and neither is the old one if the move cannot be completed.
// The return value says whether anything moved, so the caller can mention it.
func AdoptLegacy(newPath, oldPath string) (bool, error) {
	if oldPath == "" || newPath == oldPath {
		return false, nil
	}
	if _, err := os.Stat(newPath); err == nil {
		return false, nil // already have notes here; leave both alone
	}
	if _, err := os.Stat(oldPath); err != nil {
		return false, nil // nothing to adopt
	}

	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return false, fmt.Errorf("creating the new data directory: %w", err)
	}

	// Copy rather than rename: a copy that fails leaves the original intact,
	// and the two paths may be on different filesystems. The write-ahead log
	// and shared-memory files carry recent writes, so they come too.
	for _, suffix := range []string{"", "-wal", "-shm"} {
		src, dst := oldPath+suffix, newPath+suffix
		if _, err := os.Stat(src); err != nil {
			continue
		}
		if err := copyFile(src, dst); err != nil {
			// Undo a partial move so the new path stays empty and the next
			// run tries again rather than opening half a database.
			for _, s := range []string{"", "-wal", "-shm"} {
				os.Remove(newPath + s)
			}
			return false, fmt.Errorf("copying %s: %w", filepath.Base(src), err)
		}
	}
	return true, nil
}
