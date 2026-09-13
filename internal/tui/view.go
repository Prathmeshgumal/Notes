package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	switch m.mode {
	case modeHelp:
		return m.helpView()
	case modeEdit:
		return m.editView()
	default:
		return m.listView()
	}
}

func (m model) listView() string {
	paneHeight := m.height - 2
	if paneHeight < 5 {
		paneHeight = 5
	}
	inner := paneHeight - 2

	sidebar := paneStyle.
		Width(sidebarWidth).
		Height(inner).
		Render(m.sidebarContent(inner))

	header := "Preview"
	if n := m.selected(); n != nil {
		header = n.Title
	}
	previewWidth := m.width - sidebarWidth - 4
	if previewWidth < 20 {
		previewWidth = 20
	}
	preview := focusedPane.
		Width(previewWidth).
		Height(inner).
		Render(titleStyle.Render(truncate(header, previewWidth-2)) + "\n" + m.preview.View())

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, preview)

	if m.mode == modeConfirm {
		if n := m.selected(); n != nil {
			return body + "\n" + errStyle.Render(
				fmt.Sprintf(" Delete %q? ", truncate(n.Title, 40))) +
				helpStyle.Render("y / n")
		}
	}
	if m.mode == modeSearch {
		return body + "\n" + m.search.View() + "\n" + helpStyle.Render(" "+m.helpLine())
	}
	return body + "\n" + m.footer()
}

func (m model) sidebarContent(height int) string {
	if len(m.notes) == 0 {
		empty := "No notes yet."
		if m.search.Value() != "" {
			empty = "Nothing matches."
		}
		return titleStyle.Render("Notes") + "\n\n" + dimStyle.Render("  "+empty)
	}

	head := titleStyle.Render(fmt.Sprintf("Notes (%d)", len(m.notes)))
	rows := height - 2 // header + blank line
	if rows < 1 {
		rows = 1
	}

	// Keep the cursor on screen by scrolling the window of visible rows.
	start := 0
	if m.cursor >= rows {
		start = m.cursor - rows + 1
	}
	end := min(start+rows, len(m.notes))

	var b strings.Builder
	b.WriteString(head + "\n\n")
	for i := start; i < end; i++ {
		n := m.notes[i]
		label := truncate(n.Title, sidebarWidth-4)
		if i == m.cursor {
			b.WriteString(cursorStyle.Render("▸ ") + selectedStyle.Render(label))
		} else {
			b.WriteString("  " + label)
		}
		if i < end-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func (m model) editView() string {
	label := "New note"
	if m.editing != nil {
		label = "Editing · " + relativeTime(m.editing.UpdatedAt)
	}

	titleBox := paneStyle.Width(m.width - 4).Render(m.title.View())
	if m.focusTitle {
		titleBox = focusedPane.Width(m.width - 4).Render(m.title.View())
	}
	bodyBox := focusedPane.Width(m.width - 4).Render(m.body.View())
	if m.focusTitle {
		bodyBox = paneStyle.Width(m.width - 4).Render(m.body.View())
	}

	return titleStyle.Render(" "+label) + "\n" + titleBox + "\n" + bodyBox + "\n" + m.footer()
}

func (m model) helpView() string {
	return paneStyle.
		Width(m.width-4).
		Height(m.height-4).
		Render(helpText) + "\n" + helpStyle.Render(" "+m.helpLine())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
