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
		return "ctrl+s save  ctrl+b bold  alt+i italic  ctrl+k link  tab title/body  ctrl+e $EDITOR  esc cancel"
	case modeConfirm:
		return "y move to trash   n / esc cancel"
	case modeHelp:
		return "any key to close"
	default:
		return "j/k move  ↵ edit  n new  / search  o link  d trash  u undo  w web  ? help  q quit"
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
    ctrl+b       bold the word under the cursor (while editing)
    alt+i        italic — ctrl+i is Tab in a terminal, so alt is used
    ctrl+k       turn the word into a link
    click        move the cursor (while editing)
    ctrl+e       open $EDITOR (while editing)
    ctrl+s       save
    esc          cancel

  Other
    o            open a link from this note (again for the next one)
                 ctrl+click the link text works too
    d            move to trash (asks first)
    u            undo the last delete
    w            start the web UI and open a browser
    r            reload from disk
    ?            this help
    q / ctrl+c   quit
`
