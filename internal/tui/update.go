package tui

import (
	"fmt"

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

		// Notes are listed most-recently-edited first, so the list reorders as
		// you work. Follow the note itself rather than its old position.
		want := m.keepID
		if want == "" {
			if n := m.selected(); n != nil {
				want = n.ID
			}
		}
		m.keepID = ""

		m.notes = msg.notes
		if want != "" {
			for i, n := range m.notes {
				if n.ID == want {
					m.cursor = i
					break
				}
			}
		}
		if m.cursor >= len(m.notes) {
			m.cursor = max(0, len(m.notes)-1)
		}
		m.renderPreview()
		return m, nil

	case trashLoadedMsg:
		m.err = msg.err
		m.trash = msg.notes
		if m.trashCursor >= len(m.trash) {
			m.trashCursor = max(0, len(m.trash)-1)
		}
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
	case modeTrash:
		switch msg.String() {
		case "esc", "q", "T":
			m.mode = modeList
			return m, m.reload()
		case "j", "down":
			if m.trashCursor < len(m.trash)-1 {
				m.trashCursor++
			}
		case "k", "up":
			if m.trashCursor > 0 {
				m.trashCursor--
			}
		case "enter", "u", "r":
			return m, m.restoreFromTrash()
		case "d":
			if n := m.selectedTrash(); n != nil {
				m.ask(confirmPurgeNote, n.ID,
					"Delete \""+truncate(n.Title, 40)+"\" for good? This cannot be undone.")
			}
		case "E":
			if n := len(m.trash); n == 1 {
				m.ask(confirmEmptyTrash, "",
					"Permanently delete the note in the trash? This cannot be undone.")
			} else if n > 1 {
				m.ask(confirmEmptyTrash, "", fmt.Sprintf(
					"Permanently delete all %d notes in the trash? This cannot be undone.", n))
			}
		}
		return m, nil

	case modeHelp:
		m.mode = modeList
		return m, nil

	case modeConfirm:
		m.mode = m.confirmReturn
		if msg.String() == "y" || msg.String() == "Y" {
			return m, m.runConfirmed()
		}
		m.confirmKind, m.confirmID, m.confirmPrompt = confirmNone, "", ""
		return m, nil

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
		// ctrl+i is Tab in a terminal, so italic takes alt+i instead.
		if msg.String() == "alt+i" && !m.focusTitle {
			m.markBody("*", "*")
			return m, nil
		}
		switch msg.Type {
		case tea.KeyCtrlB:
			if !m.focusTitle {
				m.markBody("**", "**")
				return m, nil
			}
		case tea.KeyCtrlK:
			if !m.focusTitle {
				m.linkBody()
				return m, nil
			}
		}
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
			m.linkCursor = 0
			m.renderPreview()
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
			m.linkCursor = 0
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
		if n := m.selected(); n != nil {
			m.ask(confirmTrashNote, n.ID,
				"Move \""+truncate(n.Title, 40)+"\" to the trash?")
		}
	case "o":
		return m, m.openLink()
	case "w":
		return m, m.toggleWeb()
	case "u":
		return m, m.undo()
	case "T":
		m.trashCursor = 0
		m.mode = modeTrash
		return m, m.loadTrash()
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
