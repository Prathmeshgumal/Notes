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

	previewWidth := m.width - asideWidth - 4
	if previewWidth < 20 {
		previewWidth = 20
	}

	header := "Preview"
	if n := m.selected(); n != nil {
		header = n.Title
	}

	// The bar sits inside the pane, so the text is one column narrower than the
	// pane. It stays blank when the whole note already fits.
	scrolled := withScrollbar(
		m.preview.View(),
		m.preview.Height,
		m.preview.TotalLineCount(),
		m.preview.YOffset,
	)

	note := focusedPane.
		Width(previewWidth).
		Height(inner).
		Render(titleStyle.Render(truncate(header, previewWidth-4)) + "\n" + scrolled)

	// The right-hand column: the list of notes in a box at the top, and what is
	// known about the selected one underneath, so the corner is not dead space.
	listRows := min(asideListRows, inner-6)
	if listRows < 3 {
		listRows = 3
	}
	detailRows := inner - listRows - 2
	if detailRows < 1 {
		detailRows = 1
	}

	list := paneStyle.
		Width(asideWidth).
		Height(listRows).
		Render(m.asideList(listRows))

	details := paneStyle.
		Width(asideWidth).
		Height(detailRows).
		Render(m.asideDetails())

	aside := lipgloss.JoinVertical(lipgloss.Left, list, details)
	body := lipgloss.JoinHorizontal(lipgloss.Top, note, aside)

	if m.mode == modeSearch {
		return body + "\n" + m.search.View() + "\n" + helpStyle.Render(" "+m.helpLine())
	}
	return body + "\n" + m.footer()
}

// asideList is the note list as it appears in the top-right box.
func (m model) asideList(rows int) string {
	if len(m.notes) == 0 {
		empty := "No notes yet."
		if m.search.Value() != "" {
			empty = "Nothing matches."
		}
		return titleStyle.Render("Notes") + "\n\n" + dimStyle.Render(" "+empty)
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Notes (%d)", len(m.notes))) + "\n")

	visible := rows - 1
	if visible < 1 {
		visible = 1
	}
	// Keep the cursor in view by sliding the window of titles.
	start := 0
	if m.cursor >= visible {
		start = m.cursor - visible + 1
	}
	end := min(start+visible, len(m.notes))

	for i := start; i < end; i++ {
		label := truncate(m.notes[i].Title, asideWidth-5)
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

// asideDetails fills the corner below the list with what is worth knowing about
// the note being read.
func (m model) asideDetails() string {
	n := m.selected()
	if n == nil {
		return dimStyle.Render("Nothing selected.")
	}

	words := len(strings.Fields(n.Content))
	done, total := countTasks(n.Content)
	lines := strings.Count(n.Content, "\n") + 1

	var b strings.Builder
	b.WriteString(dimStyle.Render("Edited") + "\n")
	b.WriteString(relativeTime(n.UpdatedAt) + "\n\n")

	if total > 0 {
		b.WriteString(dimStyle.Render("Tasks") + "\n")
		b.WriteString(fmt.Sprintf("%d of %d done", done, total) + "\n\n")
	}

	b.WriteString(dimStyle.Render("Length") + "\n")
	b.WriteString(fmt.Sprintf("%d words · %d lines", words, lines))
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
