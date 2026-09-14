package tui

// The geometry of the list view, in screen cells.
//
// The renderer and the mouse handler both read it, so a click lands on what
// the eye actually sees. Working the numbers out twice — once to draw and once
// to hit-test — is how a click ends up one row off the thing it is pointing at.
type layout struct {
	inner        int // rows inside a pane's border
	previewWidth int // content width of the note pane
	asideX       int // first column of the right-hand column
	listY        int // first row of the list box, border included
	listRows     int // content rows of the list box
	itemsY       int // first row carrying a note title
	visible      int // how many titles fit
}

func (m model) geometry() layout {
	paneHeight := m.height - 2
	if paneHeight < 5 {
		paneHeight = 5
	}
	inner := paneHeight - 2

	previewWidth := m.width - asideWidth - 4
	if previewWidth < 20 {
		previewWidth = 20
	}

	listRows := inner - asideDetailRows - 2
	if listRows < 3 {
		listRows = 3
	}
	visible := listRows - 1
	if visible < 1 {
		visible = 1
	}

	return layout{
		inner:        inner,
		previewWidth: previewWidth,
		// The note pane's border adds a column on each side.
		asideX: previewWidth + 2,
		// The details box sits above, its own border included.
		listY:    asideDetailRows + 2,
		listRows: listRows,
		// Past the list box's top border, then past the "Notes (n)" heading.
		itemsY:  asideDetailRows + 4,
		visible: visible,
	}
}

// window is the slice of notes on screen. The list scrolls only far enough to
// keep the cursor visible, so the first title is not always the first note.
func (l layout) window(cursor, total int) (start, end int) {
	if cursor >= l.visible {
		start = cursor - l.visible + 1
	}
	end = min(start+l.visible, total)
	return start, end
}

// noteAt maps a click to a note, or returns -1 where there is no title.
func (m model) noteAt(x, y int) int {
	l := m.geometry()
	if x < l.asideX || y < l.itemsY {
		return -1
	}
	start, end := l.window(m.cursor, len(m.notes))
	i := start + (y - l.itemsY)
	if i < start || i >= end {
		return -1
	}
	return i
}
