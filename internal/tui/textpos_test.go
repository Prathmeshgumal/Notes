package tui

import "testing"

func TestOffsetRoundTrip(t *testing.T) {
	v := "first\nsecond line\nthird"
	for _, tc := range []struct{ row, col, off int }{
		{0, 0, 0},
		{0, 5, 5},
		{1, 0, 6},
		{1, 6, 12},
		{2, 5, 23},
	} {
		if got := offsetAt(v, tc.row, tc.col); got != tc.off {
			t.Errorf("offsetAt(%d,%d) = %d, want %d", tc.row, tc.col, got, tc.off)
		}
		r, c := rowColAt(v, tc.off)
		if r != tc.row || c != tc.col {
			t.Errorf("rowColAt(%d) = (%d,%d), want (%d,%d)", tc.off, r, c, tc.row, tc.col)
		}
	}
}

func TestOffsetClampsOutOfRange(t *testing.T) {
	v := "a\nb"
	if got := offsetAt(v, 99, 99); got != len(v) {
		t.Errorf("offsetAt past the end = %d, want %d", got, len(v))
	}
	if got := offsetAt(v, -3, -3); got != 0 {
		t.Errorf("offsetAt before the start = %d, want 0", got)
	}
	if r, c := rowColAt(v, 999); r != 1 || c != 1 {
		t.Errorf("rowColAt past the end = (%d,%d), want (1,1)", r, c)
	}
}

func TestMarkWrapsTheWordUnderTheCursor(t *testing.T) {
	got, pos := mark("make this bold", 12, "**", "**")
	if got != "make this **bold**" {
		t.Errorf("got %q", got)
	}
	if pos != len("make this **bold**") {
		t.Errorf("cursor at %d, want the end of the wrapped word", pos)
	}
}

func TestItalicOnBoldTextNestsRatherThanStripping(t *testing.T) {
	got, _ := mark("make this **bold**", 14, "*", "*")
	if got != "make this ***bold***" {
		t.Errorf("italic on bold text should nest, got %q", got)
	}
}

func TestMarkTogglesBackOff(t *testing.T) {
	wrapped, _ := mark("make this bold", 12, "**", "**")
	// Cursor inside the word again.
	unwrapped, _ := mark(wrapped, 13, "**", "**")
	if unwrapped != "make this bold" {
		t.Errorf("pressing it twice should undo it, got %q", unwrapped)
	}
}

func TestMarkOnWhitespaceInsertsAnEmptyPair(t *testing.T) {
	got, pos := mark("a  b", 2, "**", "**")
	if got != "a **** b" {
		t.Errorf("got %q", got)
	}
	if got[pos:pos+2] != "**" {
		t.Errorf("cursor should sit between the marks, got %q at %d", got, pos)
	}
}

func TestMarkKeepsTrailingPunctuationOutside(t *testing.T) {
	got, _ := mark("end of the line.", 14, "**", "**")
	if got != "end of the **line**." {
		t.Errorf("got %q", got)
	}
}

func TestLinkWrapsWordAndPositionsForTheURL(t *testing.T) {
	got, pos := link("see codehelp today", 8)
	if got != "see [codehelp]() today" {
		t.Errorf("got %q", got)
	}
	if got[pos-1] != '(' || got[pos] != ')' {
		t.Errorf("cursor should sit between the parentheses, got %q at %d", got, pos)
	}
}

func TestLinkOnEmptySpace(t *testing.T) {
	got, pos := link("a  b", 2)
	if got != "a []() b" {
		t.Errorf("got %q", got)
	}
	if got[pos-1] != '[' {
		t.Errorf("cursor should sit inside the brackets, got %q at %d", got, pos)
	}
}
