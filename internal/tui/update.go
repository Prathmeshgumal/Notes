package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
		m.renderPreview()
		return m, nil

	case reloadedMsg:
		m.err = msg.err
		m.notes = msg.notes
		if m.cursor >= len(m.notes) {
			m.cursor = max(0, len(m.notes)-1)
		}
		m.renderPreview()
		return m, nil

	case statusMsg:
		m.status = string(msg)
		return m, nil

	case clearStatusMsg:
		m.status = ""
		return m, nil

	case editorDoneMsg:
		m.body.SetValue(string(msg))
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Quitting always works, whatever the mode.
	if msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}
	m.err = nil

	switch m.mode {
	case modeHelp:
		m.mode = modeList
		return m, nil

	case modeConfirm:
		switch msg.String() {
		case "y", "Y":
			m.mode = modeList
			n := m.selected()
			if n == nil {
				return m, nil
			}
			if err := m.st.Delete(n.ID); err != nil {
				m.err = err
				return m, nil
			}
			m.lastDeleted = n.ID
			return m, tea.Batch(m.reload(), flash("Moved to trash — press u to undo"))
		default:
			m.mode = modeList
			return m, nil
		}

	case modeSearch:
		switch msg.Type {
		case tea.KeyEsc:
			m.search.SetValue("")
			m.search.Blur()
			m.mode = modeList
			return m, m.reload()
		case tea.KeyEnter:
			m.search.Blur()
			m.mode = modeList
			return m, nil
		}
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		m.cursor = 0
		return m, tea.Batch(cmd, m.reload())

	case modeEdit:
		switch msg.Type {
		case tea.KeyCtrlS:
			return m, m.save()
		case tea.KeyEsc:
			m.mode = modeList
			m.editing = nil
			m.body.Blur()
			m.title.Blur()
			return m, flash("Discarded")
		case tea.KeyCtrlE:
			return m, m.externalEdit()
		case tea.KeyTab:
			m.focusTitle = !m.focusTitle
			if m.focusTitle {
				m.body.Blur()
				m.title.Focus()
			} else {
				m.title.Blur()
				m.body.Focus()
			}
			return m, nil
		}
		var cmd tea.Cmd
		if m.focusTitle {
			m.title, cmd = m.title.Update(msg)
		} else {
			m.body, cmd = m.body.Update(msg)
		}
		return m, cmd
	}

	// List mode.
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "j", "down":
		if m.cursor < len(m.notes)-1 {
			m.cursor++
			m.renderPreview()
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
			m.renderPreview()
		}
	case "g":
		m.cursor = 0
		m.renderPreview()
	case "G":
		m.cursor = max(0, len(m.notes)-1)
		m.renderPreview()
	case "ctrl+d":
		m.preview.HalfViewDown()
	case "ctrl+u":
		m.preview.HalfViewUp()
	case "enter":
		if n := m.selected(); n != nil {
			m.startEdit(n)
		}
	case "n":
		m.startEdit(nil)
	case "e":
		if n := m.selected(); n != nil {
			m.startEdit(n)
			return m, m.externalEdit()
		}
	case "/":
		m.mode = modeSearch
		m.search.Focus()
	case "d":
		if m.selected() != nil {
			m.mode = modeConfirm
		}
	case "w":
		return m, m.toggleWeb()
	case "u":
		if m.lastDeleted == "" {
			return m, flash("Nothing to undo")
		}
		if err := m.st.Restore(m.lastDeleted); err != nil {
			m.err = err
			return m, nil
		}
		m.lastDeleted = ""
		return m, tea.Batch(m.reload(), flash("Restored"))
	case "r":
		return m, tea.Batch(m.reload(), flash("Reloaded"))
	case "?":
		m.mode = modeHelp
	}
	return m, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
