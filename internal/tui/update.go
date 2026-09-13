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

	case modeRaw:
		switch msg.String() {
		case "esc", "q", "R":
			m.mode = modeList
		case "down", "j":
			m.rawView.LineDown(1)
		case "up", "k":
			m.rawView.LineUp(1)
		case "pgdown", " ":
			m.rawView.ViewDown()
		case "pgup", "b":
			m.rawView.ViewUp()
		case "ctrl+d":
			m.rawView.HalfViewDown()
		case "ctrl+u":
			m.rawView.HalfViewUp()
		case "home", "g":
			m.rawView.GotoTop()
		case "end", "G":
			m.rawView.GotoBottom()
		case "y":
			if n := m.selected(); n != nil {
				return m, flash(copyToClipboard(n.Content))
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
		// Formatting works on the body only; the title is a single line.
		// These are alt+ combinations because the terminal spends most of the
		// control range on its own codes: ctrl+i is Tab, ctrl+h is Backspace,
		// ctrl+m is Enter, ctrl+q and ctrl+s are flow control.
		if !m.focusTitle {
			switch msg.String() {
			case "alt+i":
				m.markBody("*", "*")
				return m, nil
			case "alt+s":
				m.markBody("~~", "~~")
				return m, nil
			case "alt+c":
				m.markBody("`", "`")
				return m, nil
			case "alt+h":
				m.applyLine(heading)
				return m, nil
			case "alt+q":
				m.applyLine(quote)
				return m, nil
			// Letters, not digits: GNOME Terminal and others bind alt+1..9 to
			// switching tabs, so a digit never reaches the program. The digits
			// stay as aliases for terminals that do pass them through.
			case "alt+l", "alt+8":
				m.applyLine(bullet)
				return m, nil
			case "alt+o", "alt+7":
				m.applyLine(numbered)
				return m, nil
			case "alt+t":
				m.applyLine(task)
				return m, nil
			case "alt+x":
				m.applyLine(toggleTick)
				return m, nil
			case "alt+r":
				m.applyLine(rule)
				return m, nil
			case "alt+f":
				m.applyLine(codeBlock)
				return m, nil
			}
		}
		// An alt+key that reached here is not a formatting action. Swallow it
		// rather than letting the field insert the bare rune, which would type
		// an "8" for alt+8.
		if msg.Alt && msg.Type == tea.KeyRunes {
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
		case tea.KeyCtrlY:
			return m, flash(copyToClipboard(m.body.Value()))
		case tea.KeyCtrlP:
			// Preview what is being written, the Write/Preview pair the web
			// editor has. The draft is rendered, not the saved note.
			m.previewDraft = !m.previewDraft
			if m.previewDraft {
				m.renderDraft()
			}
			return m, nil
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
	case "j":
		m.moveCursor(1)
	case "k":
		m.moveCursor(-1)
	case "g":
		m.cursor = 0
		m.renderPreview()
	case "G":
		m.cursor = max(0, len(m.notes)-1)
		m.renderPreview()

	// The preview scrolls; the list does not. A terminal turns the mouse wheel
	// into arrow keys in the alternate screen, so binding the arrows here is
	// what makes the wheel scroll the note instead of jumping between notes.
	case "down":
		m.preview.LineDown(1)
	case "up":
		m.preview.LineUp(1)
	case "pgdown", " ":
		m.preview.ViewDown()
	case "pgup", "b":
		m.preview.ViewUp()
	case "ctrl+d":
		m.preview.HalfViewDown()
	case "ctrl+u":
		m.preview.HalfViewUp()
	case "home":
		m.preview.GotoTop()
	case "end":
		m.preview.GotoBottom()

	case "y":
		if n := m.selected(); n != nil {
			return m, flash(copyToClipboard(n.Content))
		}
	case "R":
		if n := m.selected(); n != nil {
			m.mode = modeRaw
			m.rawView.SetContent(n.Content)
			m.rawView.GotoTop()
		}
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
