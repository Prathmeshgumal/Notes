package tui

type mode int

const (
	modeList mode = iota
	modeSearch
	modeEdit
	modeConfirm
	modeHelp
)

// helpLine is the context-sensitive hint bar along the bottom.
func (m model) helpLine() string {
	switch m.mode {
	case modeSearch:
		return "type to filter   ↵ accept   esc clear"
	case modeEdit:
		return "ctrl+s save   tab title/body   ctrl+e $EDITOR   esc cancel"
	case modeConfirm:
		return "y delete   n / esc cancel"
	case modeHelp:
		return "any key to close"
	default:
		return "j/k move   ↵ edit   n new   / search   d delete   w web   ? help   q quit"
	}
}

const helpText = `
  Notes — keys

  Moving
    j / ↓        next note
    k / ↑        previous note
    g / G        first / last note
    /            search, esc to clear

  Writing
    ↵            edit the selected note
    n            new note
    ctrl+e       open $EDITOR (while editing)
    ctrl+s       save
    esc          cancel

  Other
    d            delete (asks first)
    w            start the web UI and open a browser
    r            reload from disk
    ?            this help
    q / ctrl+c   quit
`
