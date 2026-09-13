package tui

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/prathmesh/notes/internal/store"
)

func benchModel(b *testing.B) model {
	b.Helper()
	st, err := store.Open(filepath.Join(b.TempDir(), "bench.db"))
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { st.Close() })
	body := "# Heading\n\nSome **bold** text with a [link](https://example.com).\n\n" +
		"- one\n- two\n- three\n\n> a quote\n\n```go\nfmt.Println(\"hi\")\n```\n"
	for _, t := range []string{"alpha", "beta", "gamma"} {
		if _, err := st.Create(t, body); err != nil {
			b.Fatal(err)
		}
	}
	m := New(st)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 118, Height: 55})
	return next.(model)
}

// BenchmarkCursorMove measures one keypress of navigation, which is what the
// user feels when holding down an arrow key.
func BenchmarkCursorMove(b *testing.B) {
	m := benchModel(b)
	down := tea.KeyMsg{Type: tea.KeyDown}
	up := tea.KeyMsg{Type: tea.KeyUp}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			m = press(m, down)
		} else {
			m = press(m, up)
		}
	}
}

func BenchmarkRenderPreview(b *testing.B) {
	m := benchModel(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.renderPreview()
	}
}
