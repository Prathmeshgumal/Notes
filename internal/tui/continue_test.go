package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestContinuation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		line    string
		prefix  string
		endList bool
	}{
		{"bullet", "- buy milk", "- ", false},
		{"star bullet", "* buy milk", "* ", false},
		{"plus bullet", "+ buy milk", "+ ", false},
		{"task", "- [ ] buy milk", "- [ ] ", false},
		{"finished task starts a fresh one", "- [x] buy milk", "- [ ] ", false},
		{"numbered", "1. first", "2. ", false},
		{"numbered keeps counting", "9. ninth", "10. ", false},
		{"numbered with a paren", "3) third", "4) ", false},
		{"quote", "> a thought", "> ", false},
		{"nested quote", "> > deeper", "> > ", false},

		{"indented bullet keeps its indent", "   - nested", "   - ", false},
		{"indented task", "   - [ ] nested", "   - [ ] ", false},
		{"indented numbered", "  2. second", "  3. ", false},

		{"empty bullet ends the list", "- ", "", true},
		{"empty task ends the list", "- [ ] ", "", true},
		{"empty numbered ends the list", "1. ", "", true},
		{"empty quote ends it", "> ", "", true},

		{"ordinary prose", "just a sentence", "", false},
		{"heading", "# A heading", "", false},
		{"empty line", "", "", false},
		{"a rule is not a list", "---", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prefix, end := continuation(tc.line)
			if prefix != tc.prefix || end != tc.endList {
				t.Errorf("continuation(%q) = (%q, %v), want (%q, %v)",
					tc.line, prefix, end, tc.prefix, tc.endList)
			}
		})
	}
}

// The whole point is what happens when Enter is actually pressed.
func TestEnterCarriesListsOn(t *testing.T) {
	for _, tc := range []struct {
		name  string
		typed string
		want  string
	}{
		{"task list", "- [ ] first", "- [ ] first\n- [ ] "},
		{"bullet list", "- first", "- first\n- "},
		{"numbered list", "1. first", "1. first\n2. "},
		{"quote", "> first", "> first\n> "},
		{"indented task", "   - [ ] first", "   - [ ] first\n   - [ ] "},
		{"prose is untouched", "just a sentence", "just a sentence\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, _ := newTestModel(t)
			m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
			m = press(m, key('n'))
			for _, r := range tc.typed {
				m = press(m, key(r))
			}
			m = press(m, tea.KeyMsg{Type: tea.KeyEnter})
			if got := m.body.Value(); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// Typing straight after Enter continues the item, rather than landing before
// the marker.
func TestTypingAfterEnterLandsInTheNewItem(t *testing.T) {
	m, _ := newTestModel(t)
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, key('n'))
	for _, r := range "- [ ] milk" {
		m = press(m, key(r))
	}
	m = press(m, tea.KeyMsg{Type: tea.KeyEnter})
	for _, r := range "eggs" {
		m = press(m, key(r))
	}
	if got := m.body.Value(); got != "- [ ] milk\n- [ ] eggs" {
		t.Errorf("got %q", got)
	}
}

// Enter on an item with nothing in it ends the list.
func TestEnterOnAnEmptyItemEndsTheList(t *testing.T) {
	m, _ := newTestModel(t)
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, key('n'))
	for _, r := range "- [ ] milk" {
		m = press(m, key(r))
	}
	m = press(m, tea.KeyMsg{Type: tea.KeyEnter}) // gives "- [ ] "
	m = press(m, tea.KeyMsg{Type: tea.KeyEnter}) // the item is empty: end it

	if got := m.body.Value(); got != "- [ ] milk\n" {
		t.Errorf("got %q, want the empty item cleared", got)
	}
	for _, r := range "a plain line" {
		m = press(m, key(r))
	}
	if got := m.body.Value(); got != "- [ ] milk\na plain line" {
		t.Errorf("after ending the list, got %q", got)
	}
}

// alt+enter breaks the line without carrying the list on.
func TestAltEnterSkipsTheContinuation(t *testing.T) {
	m, _ := newTestModel(t)
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, key('n'))
	for _, r := range "- [ ] milk" {
		m = press(m, key(r))
	}
	m = press(m, tea.KeyMsg{Type: tea.KeyEnter, Alt: true})
	if got := m.body.Value(); got != "- [ ] milk\n" {
		t.Errorf("got %q, want a plain line break", got)
	}
}

// A numbered list keeps counting across several presses.
func TestNumberedListKeepsCounting(t *testing.T) {
	m, _ := newTestModel(t)
	m = press(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(m, key('n'))
	for _, r := range "1. one" {
		m = press(m, key(r))
	}
	for _, word := range []string{"two", "three"} {
		m = press(m, tea.KeyMsg{Type: tea.KeyEnter})
		for _, r := range word {
			m = press(m, key(r))
		}
	}
	if got := m.body.Value(); got != "1. one\n2. two\n3. three" {
		t.Errorf("got %q", got)
	}
}
