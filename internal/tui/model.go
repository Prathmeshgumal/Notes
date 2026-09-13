// Package tui is the terminal interface: a two-pane, keyboard-driven view of
// the notes, in the spirit of lazygit.
package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/prathmesh/notes/internal/store"
	"github.com/prathmesh/notes/internal/web"
)

const sidebarWidth = 30

type model struct {
	st    *store.Store
	notes []store.Note

	cursor int
	mode   mode
	width  int
	height int

	search  textinput.Model
	title   textinput.Model
	body    textarea.Model
	preview viewport.Model

	editing    *store.Note // nil while composing a brand-new note
	focusTitle bool

	server *web.Server
	status string
	err    error

	lastDeleted string // id of the most recent delete, for undo

	// Rendering Markdown is the expensive part of moving the cursor, so the
	// renderer is built once per width and the output cached per note.
	renderer      *glamour.TermRenderer
	rendererWidth int
	glamourStyle  string
	rendered      map[string]string
}

type reloadedMsg struct {
	notes []store.Note
	err   error
}

type statusMsg string

type clearStatusMsg struct{}

func New(st *store.Store) model {
	search := textinput.New()
	search.Prompt = "/"
	search.Placeholder = "search"

	title := textinput.New()
	title.Prompt = ""
	title.Placeholder = "Title (blank uses the first line)"
	title.CharLimit = 200

	body := textarea.New()
	body.Placeholder = "Write in Markdown…"
	body.ShowLineNumbers = false
	body.CharLimit = 0

	// Ask the terminal about its background exactly once. Doing this per
	// render (via glamour's auto style) stalls every keypress.
	glamourStyle := "dark"
	if !lipgloss.HasDarkBackground() {
		glamourStyle = "light"
	}

	m := model{
		st:           st,
		glamourStyle: glamourStyle,
		rendered:     map[string]string{},
		search:       search,
		title:        title,
		body:         body,
		preview:      viewport.New(0, 0),
		mode:         modeList,
		// Sensible defaults so the first frame renders even if the terminal
		// never reports its size; WindowSizeMsg overrides these.
		width:  80,
		height: 24,
	}
	m.layout()
	// Load synchronously so the first frame already shows the notes rather
	// than flashing an empty list.
	if notes, err := st.List(""); err == nil {
		m.notes = notes
	}
	m.renderPreview()
	return m
}

func (m model) Init() tea.Cmd { return m.reload() }

func (m model) reload() tea.Cmd {
	q := m.search.Value()
	return func() tea.Msg {
		notes, err := m.st.List(q)
		return reloadedMsg{notes: notes, err: err}
	}
}

func flash(s string) tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return statusMsg(s) },
		tea.Tick(3*time.Second, func(time.Time) tea.Msg { return clearStatusMsg{} }),
	)
}

func (m *model) selected() *store.Note {
	if m.cursor < 0 || m.cursor >= len(m.notes) {
		return nil
	}
	return &m.notes[m.cursor]
}

func (m *model) layout() {
	if m.width == 0 {
		return
	}
	paneHeight := m.height - 2 // status + help lines
	if paneHeight < 3 {
		paneHeight = 3
	}
	previewWidth := m.width - sidebarWidth - 4
	if previewWidth < 20 {
		previewWidth = 20
	}
	m.preview.Width = previewWidth
	m.preview.Height = paneHeight - 3

	m.title.Width = m.width - 6
	m.body.SetWidth(m.width - 6)
	m.body.SetHeight(paneHeight - 4)
	m.search.Width = m.width - 6
}

// renderPreview turns the selected note's Markdown into styled terminal output.
// Results are cached: moving the cursor must not re-render or re-detect colours.
func (m *model) renderPreview() {
	n := m.selected()
	if n == nil {
		m.preview.SetContent(dimStyle.Render("\n  No note selected."))
		return
	}
	width := m.preview.Width
	if width < 20 {
		width = 20
	}

	if cached, ok := m.rendered[n.ID+"\x00"+n.UpdatedAt]; ok {
		m.preview.SetContent(cached)
		m.preview.GotoTop()
		return
	}

	if m.renderer == nil || m.rendererWidth != width {
		r, err := glamour.NewTermRenderer(
			glamour.WithStandardStyle(m.glamourStyle),
			glamour.WithWordWrap(width-2),
		)
		if err != nil {
			m.preview.SetContent(n.Content)
			return
		}
		m.renderer = r
		m.rendererWidth = width
		// Cached output was wrapped for the old width.
		m.rendered = map[string]string{}
	}

	out, err := m.renderer.Render(separateListGroups(n.Content))
	if err != nil {
		m.preview.SetContent(n.Content)
		return
	}
	m.rendered[n.ID+"\x00"+n.UpdatedAt] = out
	m.preview.SetContent(out)
	m.preview.GotoTop()
}

func (m *model) startEdit(n *store.Note) {
	m.mode = modeEdit
	m.focusTitle = false
	if n == nil {
		m.editing = nil
		m.title.SetValue("")
		m.body.SetValue("")
	} else {
		cp := *n
		m.editing = &cp
		m.title.SetValue(n.Title)
		m.body.SetValue(n.Content)
	}
	m.title.Blur()
	m.body.Focus()
	m.body.CursorEnd()
}

func (m *model) save() tea.Cmd {
	title, content := m.title.Value(), m.body.Value()
	if strings.TrimSpace(title) == "" && strings.TrimSpace(content) == "" {
		m.mode = modeList
		return flash("Empty note discarded")
	}
	var err error
	if m.editing == nil {
		_, err = m.st.Create(title, content)
	} else {
		_, err = m.st.Update(m.editing.ID, title, content)
	}
	if err != nil {
		m.err = err
		return nil
	}
	m.mode = modeList
	m.editing = nil
	return tea.Batch(m.reload(), flash("Saved"))
}

// externalEdit suspends the TUI, hands the body to $EDITOR, and reads it back.
func (m *model) externalEdit() tea.Cmd {
	editor := firstNonEmpty(os.Getenv("VISUAL"), os.Getenv("EDITOR"), fallbackEditor())
	if editor == "" {
		return flash("No $EDITOR set and no fallback editor found")
	}
	tmp, err := os.CreateTemp("", "note-*.md")
	if err != nil {
		m.err = err
		return nil
	}
	path := tmp.Name()
	if _, err := tmp.WriteString(m.body.Value()); err != nil {
		tmp.Close()
		m.err = err
		return nil
	}
	tmp.Close()

	parts := strings.Fields(editor)
	cmd := exec.Command(parts[0], append(parts[1:], path)...)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		defer os.Remove(path)
		if err != nil {
			return statusMsg("Editor exited: " + err.Error())
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return statusMsg("Could not read the file back")
		}
		return editorDoneMsg(string(data))
	})
}

type editorDoneMsg string

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func fallbackEditor() string {
	for _, e := range []string{"nvim", "vim", "nano", "vi"} {
		if p, err := exec.LookPath(e); err == nil {
			return p
		}
	}
	return ""
}

func (m *model) toggleWeb() tea.Cmd {
	if m.server != nil {
		_ = m.server.Stop()
		m.server = nil
		return flash("Web UI stopped")
	}
	srv, err := web.New(m.st, 4321)
	if err != nil {
		m.err = err
		return nil
	}
	srv.Start()
	m.server = srv
	openBrowser(srv.URL)
	return flash("Web UI at " + srv.URL)
}

func openBrowser(url string) {
	var cmd string
	switch {
	case commandExists("xdg-open"):
		cmd = "xdg-open"
	case commandExists("open"):
		cmd = "open"
	default:
		return
	}
	_ = exec.Command(cmd, url).Start()
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// DefaultPath puts the database next to the user's other application data.
func DefaultPath() string {
	if p := os.Getenv("NOTES_DB"); p != "" {
		return p
	}
	dir, err := os.UserHomeDir()
	if err != nil {
		return "notes.db"
	}
	return filepath.Join(dir, ".local", "share", "notes", "notes.db")
}

func (m model) footer() string {
	status := m.status
	if m.err != nil {
		status = errStyle.Render("error: " + m.err.Error())
	}
	if status == "" && m.server != nil {
		status = dimStyle.Render("web: " + m.server.URL)
	}
	return lipgloss.NewStyle().Width(m.width).Render(status) + "\n" +
		helpStyle.Width(m.width).Render(" "+m.helpLine())
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return string(r[:n])
	}
	return string(r[:n-1]) + "…"
}

func relativeTime(iso string) string {
	t, err := time.Parse(time.RFC3339Nano, iso)
	if err != nil {
		return ""
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
