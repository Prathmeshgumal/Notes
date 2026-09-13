package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	switch m.mode {
	case modeRaw:
		// Deliberately bare: no panes, no borders, no padding, so selecting
		// these lines with the mouse yields exactly the note's Markdown.
		return m.rawView.View() + "\n" + helpStyle.Render(m.helpLine())
	case modeConfirm:
		// Draw whatever is being asked about, with the question in the footer.
		if m.confirmReturn == modeTrash {
			return m.trashView()
		}
		return m.listView()
	case modeTrash:
		return m.trashView()
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
	// The bar sits inside the pane, so the text is one column narrower than the
	// pane. It stays blank when the whole note already fits.
	scrolled := withScrollbar(
		m.preview.View(),
		m.preview.Height,
		m.preview.TotalLineCount(),
		m.preview.YOffset,
	)

	preview := focusedPane.
		Width(previewWidth).
		Height(inner).
		Render(titleStyle.Render(truncate(header, previewWidth-4)) + "\n" + scrolled)

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, preview)

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
	label := titleStyle.Render("New note")
	if m.editing != nil {
		label = titleStyle.Render("Editing") + dimStyle.Render(" · "+relativeTime(m.editing.UpdatedAt))
	}

	titleBox := paneStyle.Width(m.width - 4).Render(m.title.View())
	if m.focusTitle {
		titleBox = focusedPane.Width(m.width - 4).Render(m.title.View())
	}

	if m.previewDraft {
		label += dimStyle.Render("  ·  preview")
		box := focusedPane.Width(m.width - 4).Render(m.draft.View())
		return " " + label + "\n" + titleBox + "\n" + box + "\n" + m.footer()
	}

	bodyBox := focusedPane.Width(m.width - 4).Render(m.body.View())
	if m.focusTitle {
		bodyBox = paneStyle.Width(m.width - 4).Render(m.body.View())
	}

	return " " + label + "\n" + titleBox + "\n" + bodyBox + "\n" + m.footer()
}

func (m model) trashView() string {
	inner := m.height - 4
	if inner < 3 {
		inner = 3
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Trash (%d)", len(m.trash))) + "\n")
	b.WriteString(dimStyle.Render("Deleted notes are kept for 30 days.") + "\n\n")

	if len(m.trash) == 0 {
		b.WriteString(dimStyle.Render("  The trash is empty."))
	} else {
		rows := inner - 3
		if rows < 1 {
			rows = 1
		}
		start := 0
		if m.trashCursor >= rows {
			start = m.trashCursor - rows + 1
		}
		end := min(start+rows, len(m.trash))

		for i := start; i < end; i++ {
			n := m.trash[i]
			label := truncate(n.Title, m.width-24)
			line := "  " + label
			if i == m.trashCursor {
				line = cursorStyle.Render("▸ ") + selectedStyle.Render(label)
			}
			b.WriteString(line + "\n")
		}
	}

	return paneStyle.Width(m.width-4).Height(inner).Render(b.String()) +
		"\n" + m.footer()
}

func (m model) helpView() string {
	return paneStyle.
		Width(m.width-4).
		Height(m.height-4).
		Render(m.help.View()) + "\n" + helpStyle.Render(" "+m.helpLine())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
