package tui

type mode int

const (
	modeList mode = iota
	modeSearch
	modeEdit
	modeConfirm
	modeHelp
	modeTrash
)

// helpLine is the context-sensitive hint bar along the bottom.
func (m model) helpLine() string {
	switch m.mode {
	case modeSearch:
		return "type to filter   ↵ accept   esc clear"
	case modeEdit:
		if m.previewDraft {
			return "ctrl+p back to writing   ctrl+s save   esc cancel"
		}
		return "ctrl+s save  ctrl+p preview  ctrl+b bold  alt+i italic  ctrl+k link  ? in help: all keys  esc cancel"
	case modeConfirm:
		return "y confirm   n / esc cancel"
	case modeTrash:
		return "j/k move   ↵ restore   d delete for good   E empty trash   esc back"
	case modeHelp:
		return "any key to close"
	default:
		return "j/k move  ↵ edit  n new  / search  o link  d trash  u undo  T trash  w web  ? help  q quit"
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
    ctrl+s       save
    ctrl+p       preview what you are writing
    tab          switch between the title and the body
    ctrl+e       open $EDITOR
    esc          cancel

  Formatting, while editing the body
    ctrl+b       bold          alt+8    bulleted list
    alt+i        italic        alt+7    numbered list
    alt+s        strikethrough alt+t    task list
    alt+c        inline code   alt+x    tick / untick a task
    alt+f        code block    alt+r    horizontal rule
    ctrl+k       link          alt+h    heading (press again for deeper)
    alt+q        blockquote

    Most are alt+ because a terminal spends the control range on its own
    codes: ctrl+i is Tab, ctrl+h is Backspace, ctrl+m is Enter.

  Other
    o            open a link from this note (again for the next one)
                 ctrl+click the link text works too
    d            move to trash (asks first)
    u            undo the last delete (again for the one before it)
    T            the trash — restore anything deleted in the last 30 days
                 inside it: ↵ restore, d delete for good, E empty the trash
    w            start the web UI and open a browser
    r            reload from disk
    ?            this help
    q / ctrl+c   quit
`
