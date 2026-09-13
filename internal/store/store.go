// Package store is the single source of truth for notes. Both the terminal UI
// and the web server talk to it; it owns a plain SQLite file on disk.
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // pure-Go driver: no cgo, no system libraries
)

type Note struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Store struct {
	db   *sql.DB
	Path string
}

var ErrNotFound = errors.New("note not found")

const schema = `
CREATE TABLE IF NOT EXISTS notes (
  id         TEXT PRIMARY KEY,
  title      TEXT NOT NULL DEFAULT 'Untitled',
  content    TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS notes_updated_at_idx ON notes (updated_at DESC);`

// Open prepares the database file, creating parent directories as needed.
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}
	// WAL lets the TUI and the web server use the file at the same time.
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("creating schema: %w", err)
	}
	return &Store{db: db, Path: path}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func now() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func scan(rows *sql.Rows) ([]Note, error) {
	defer rows.Close()
	notes := []Note{}
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

// List returns notes newest-first, optionally filtered by a substring match on
// the title or body. SQLite's LIKE is already case-insensitive for ASCII.
func (s *Store) List(query string) ([]Note, error) {
	const cols = `SELECT id, title, content, created_at, updated_at FROM notes`
	if q := strings.TrimSpace(query); q != "" {
		like := "%" + q + "%"
		rows, err := s.db.Query(cols+` WHERE title LIKE ? OR content LIKE ?
			ORDER BY updated_at DESC`, like, like)
		if err != nil {
			return nil, err
		}
		return scan(rows)
	}
	rows, err := s.db.Query(cols + ` ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	return scan(rows)
}

func (s *Store) Get(id string) (Note, error) {
	var n Note
	err := s.db.QueryRow(
		`SELECT id, title, content, created_at, updated_at FROM notes WHERE id = ?`, id,
	).Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return n, ErrNotFound
	}
	return n, err
}

func (s *Store) Create(title, content string) (Note, error) {
	n := Note{
		ID:        newID(),
		Title:     DeriveTitle(title, content),
		Content:   content,
		CreatedAt: now(),
		UpdatedAt: now(),
	}
	_, err := s.db.Exec(
		`INSERT INTO notes (id, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		n.ID, n.Title, n.Content, n.CreatedAt, n.UpdatedAt)
	return n, err
}

func (s *Store) Update(id, title, content string) (Note, error) {
	res, err := s.db.Exec(`UPDATE notes SET title = ?, content = ?, updated_at = ? WHERE id = ?`,
		DeriveTitle(title, content), content, now(), id)
	if err != nil {
		return Note{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return Note{}, ErrNotFound
	}
	return s.Get(id)
}

func (s *Store) Delete(id string) error {
	res, err := s.db.Exec(`DELETE FROM notes WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) Count() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT count(*) FROM notes`).Scan(&n)
	return n, err
}

// DeriveTitle falls back to the first meaningful line of the body, gist-style,
// when no explicit title was given.
func DeriveTitle(title, content string) string {
	if t := strings.TrimSpace(title); t != "" {
		return truncate(t, 200)
	}
	for _, line := range strings.Split(content, "\n") {
		if t := strings.TrimSpace(strings.TrimLeft(line, "# ")); t != "" {
			return truncate(t, 200)
		}
	}
	return "Untitled"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
