package tui

import "strings"

// scrollbar draws a vertical position indicator for a viewport: a track the
// height of the pane with a thumb whose size and position reflect how much of
// the note is on screen and where. Returns an empty column when everything
// already fits, so short notes are not decorated with a full-height bar.
func scrollbar(height, total, offset int) []string {
	if height <= 0 {
		return nil
	}
	if total <= height {
		return make([]string, height) // nothing to scroll: a blank column
	}

	// Thumb size in proportion to how much is visible, at least one cell.
	thumb := height * height / total
	if thumb < 1 {
		thumb = 1
	}

	// Position in proportion to how far down we are. The last line of the note
	// must put the thumb at the bottom, so scale against the scrollable range.
	maxOffset := total - height
	top := 0
	if maxOffset > 0 {
		top = offset * (height - thumb) / maxOffset
	}
	if top > height-thumb {
		top = height - thumb
	}

	col := make([]string, height)
	for i := range col {
		if i >= top && i < top+thumb {
			col[i] = scrollThumbStyle.Render("┃")
		} else {
			col[i] = scrollTrackStyle.Render("│")
		}
	}
	return col
}

// withScrollbar puts the bar down the right-hand edge of a rendered pane.
func withScrollbar(content string, height, total, offset int) string {
	bar := scrollbar(height, total, offset)
	if bar == nil {
		return content
	}
	lines := strings.Split(content, "\n")
	for i := 0; i < height; i++ {
		if i < len(lines) {
			lines[i] += bar[i]
		} else {
			lines = append(lines, bar[i])
		}
	}
	return strings.Join(lines, "\n")
}
