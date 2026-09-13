package tui

type mode int

const (
	modeList mode = iota
	modeSearch
	modeEdit
	modeConfirm
	modeHelp
	modeTrash
	modeRaw
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
	case modeRaw:
		return "select with the mouse to copy   ↑/↓ scroll   y copy it all   esc back"
	case modeTrash:
		return "j/k move   ↵ restore   d delete for good   E empty trash   esc back"
	case modeHelp:
		return "any key to close"
	default:
		return "j/k note  ↑/↓ scroll  ↵ edit  n new  / search  y copy  o link  d trash  w web  ? help  q quit"
	}
}

const helpText = `
  Notes — keys

  Choosing a note
    j            next note
    k            previous note
    g / G        first / last note
    /            search, esc to clear

  Reading a long note
    ↑ / ↓        scroll a line — the mouse wheel sends these too
    space / b    scroll a page          pgup / pgdn  the same
    ctrl+d / u   scroll half a page
    home / end   jump to the top / bottom

  Copying
    y            copy the note's Markdown to the clipboard
    R            show the Markdown source full-screen, with no borders, so a
                 mouse selection copies clean Markdown and nothing else
    ctrl+y       copy while editing

  Writing
    ↵            edit the selected note
    n            new note
    ↵            new line — inside a list it starts the next item
                 press it on an empty item to end the list
    alt+↵        a plain line break, without continuing the list
    ctrl+s       save
    ctrl+p       preview what you are writing
    tab          switch between the title and the body
    ctrl+e       open $EDITOR
    esc          cancel

  Formatting, while editing the body
    ctrl+b       bold          alt+l    bulleted list
    alt+i        italic        alt+o    ordered (numbered) list
    alt+s        strikethrough alt+t    task list
    alt+c        inline code   alt+x    tick / untick a task
    alt+f        code block    alt+r    horizontal rule
    ctrl+k       link          alt+h    heading (press again for deeper)
    alt+q        blockquote

    Most are alt+ because a terminal spends the control range on its own
    codes: ctrl+i is Tab, ctrl+h is Backspace, ctrl+m is Enter. They are
    letters rather than digits because terminals bind alt+1..9 to switching
    tabs, so a digit never reaches a program running inside one.

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
