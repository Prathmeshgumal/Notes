# Using note

Everything this app can do, and how to do it.

If you have not installed it yet, the quickest way is to download the binary:

```bash
curl -L https://github.com/Prathmeshgumal/Notes/releases/latest/download/note-linux-amd64 -o note
chmod +x note && mv note ~/.local/bin/
```

There are two interfaces over the same notes: a terminal UI (the default) and a
web UI you start when you want it. Both read and write the same SQLite file, and
both can be open at once.

---

## Starting it

| Command | What it does |
| --- | --- |
| `note` | Open the terminal UI |
| `note --web` | Run only the web UI, no terminal interface |
| `note --web --port 9000` | Same, on a port of your choosing (default 4321) |
| `note --db /tmp/scratch.db` | Use a different database — handy for experimenting |
| `note --version` | Print the version and exit |

`NOTES_DB=/path/to.db note` does the same as `--db`.

Nothing runs in the background. When you quit, nothing is left running.

The first time you run it, a welcome note is written so you have something to
look at and something to try. Delete it whenever you like — nothing depends on
it, and it is never written again.

---

## The terminal UI

Two panes: your notes on the left, the selected note rendered on the right. The
bar along the bottom always shows the keys available right now.

### Moving around

| Key | Does |
| --- | --- |
| `j` | Next note |
| `k` | Previous note |
| `g` / `G` | Jump to the first / last note |
| `r` | Reload from disk (picks up changes made in the web UI) |
| `?` | Full help — any key closes it |
| `q` | Quit |

### Reading a long note

The right-hand pane scrolls, and shows a scrollbar when there is more than fits.

| Key | Does |
| --- | --- |
| `↑` / `↓` | Scroll one line |
| `space` / `b` | Scroll a page — `pgup` / `pgdn` do the same |
| `ctrl+d` / `ctrl+u` | Scroll half a page |
| `home` / `end` | Jump to the top / bottom |

**The mouse wheel scrolls the note.** A terminal turns the wheel into arrow
keys, which is why the arrows scroll the preview rather than moving between
notes — use `j` and `k` for that.

### Copying a note

| Key | Does |
| --- | --- |
| `R` | Show the Markdown source full-screen, to select with the mouse |

Press `R` and the note's Markdown fills the screen with no borders, no padding
and no scrollbar, so dragging over it selects the text and nothing else. Copy
with your terminal's own shortcut — **`Ctrl+Shift+C`** in GNOME Terminal and
most Linux terminals, `Cmd+C` on macOS. `esc` goes back.

You choose what to copy, which a key that copies the whole note cannot do, and
it works everywhere because it is your terminal doing the copying rather than
the app asking for a clipboard it may not be able to reach.

The app never captures the mouse, so selection always belongs to your terminal.
The reason `R` exists at all is that the normal view draws borders around the
text, and a selection there would carry those along with it.

Notes are listed **most recently edited first**, so whatever you last worked on
is at the top. Saving a note moves it to the top, and your selection follows it
rather than staying at that position.

### Writing

| Key | Does |
| --- | --- |
| `n` | New note |
| `↵` | Edit the selected note |
| `e` | Edit the selected note straight in `$EDITOR` |

While editing:

| Key | Does |
| --- | --- |
| `ctrl+s` | Save and go back to the list |
| `esc` | Discard changes |
| `tab` | Switch between the title field and the body |
| `ctrl+p` | Preview what you are writing; press again to go back |
| `ctrl+b` | Bold |
| `alt+i` | Italic |
| `alt+s` | Strikethrough |
| `alt+c` | Inline code |
| `alt+f` | Code block |
| `ctrl+k` | Link, cursor ready for the address |
| `alt+h` | Heading — press again for a deeper one |
| `alt+q` | Blockquote |
| `alt+l` | Bulleted list |
| `alt+o` | Ordered (numbered) list |
| `alt+t` | Task list |
| `alt+x` | Tick or untick the task under the cursor |
| `alt+r` | Horizontal rule |
| `ctrl+e` | Hand the body to `$EDITOR`; save and quit there to come back |
| `↵` | New line — inside a list it starts the next item |
| `alt+↵` | A plain line break, without continuing the list |

Notes are not limited in length. Paste a whole Markdown file in and keep
editing it.

### Lists carry on by themselves

Press `↵` at the end of a list item and the next one starts for you:

| You have | `↵` gives you |
| --- | --- |
| `- [ ] buy milk` | `- [ ] ` |
| `- [x] buy milk` | `- [ ] ` — a new task starts unticked |
| `- buy milk` | `- ` |
| `1. first` | `2. ` — and keeps counting |
| `> a thought` | `> ` |

Indentation is kept, so a nested item stays nested.

**To finish a list**, press `↵` again on the empty item it just made: the
marker is removed and you are left on a plain line.

**For a line break inside an item**, use `alt+↵`. Shift+Enter would be the
obvious choice, but a terminal sends the very same byte for it as for Enter —
they are literally indistinguishable to any program running inside one.

`ctrl+b`, `alt+i`, `alt+s`, `alt+c` and `ctrl+k` act on the word under the
cursor; the rest act on the line. All of them toggle off if you press them
again, and the list types convert between each other rather than stacking.

Most are `alt+` because a terminal spends the control range on its own codes:
**`ctrl+i` is Tab**, `ctrl+h` is Backspace and `ctrl+m` is Enter — the same
bytes, indistinguishable to any program.

They are letters rather than digits because **terminals bind `alt+1`–`alt+9` to
switching tabs**, so a digit is swallowed before any program inside can see it.
`alt+7` and `alt+8` still work as aliases in terminals that do pass them
through.

The mouse is left to your terminal, so selecting and copying text works exactly
as it does anywhere else. Move the cursor with the arrow keys.

**Titles are optional.** Leave the title blank and the first line of the note
becomes its title, the way GitHub Gists work. If that first line is also the
start of your note, the preview shows it once — as the heading — not twice.

`$EDITOR` falls back to `$VISUAL`, then to nvim, vim, nano or vi, whichever is
installed.

### Finding notes

| Key | Does |
| --- | --- |
| `/` | Start searching — the list narrows as you type |
| `↵` | Keep the filter and go back to the list |
| `esc` | Clear the search and show everything |

Search matches both titles and note bodies, and ignores case.

### Links

| Key | Does |
| --- | --- |
| `o` | Open a link from this note in your browser |

Write links as `[some text](https://example.com)`. The preview shows only
*some text* — the URL stays hidden, like it would in a browser.

**Ctrl+click the text to open it.** The preview emits real terminal hyperlinks
(OSC 8), so the link text is clickable in GNOME Terminal, iTerm2, kitty,
WezTerm, Windows Terminal and most other modern terminals.

`o` does the same thing from the keyboard, and is the fallback in a terminal
that doesn't support hyperlinks. If the note has several links, press `o` again
for the next one; the status line names the one it opened.

URLs written out in full are clickable as well — they have no text to hide
behind, so they stay visible. Link syntax inside a code block stays literal.

### Deleting, and undoing it

| Key | Does |
| --- | --- |
| `d` | Move the note to the trash — asks `y` / `n` first |
| `u` | Undo the last delete; press again to walk further back |
| `T` | Open the trash and restore anything in it |

Nothing is destroyed immediately. A trashed note stays recoverable for **30
days**, then is purged the next time the app starts.

`u` walks back through your deletes one at a time, so several deletes take
several undos.

`T` opens the trash:

| Key | Does |
| --- | --- |
| `j` / `k` | Move through the trashed notes |
| `↵` | Restore the selected note |
| `d` | Delete it for good — asks first, and cannot be undone |
| `E` | Empty the trash — asks first, and cannot be undone |
| `esc` | Back to your notes |

`d` and `E` are the only two actions in the app that destroy anything. Both name
what they are about to remove and both say plainly that it cannot be undone. A
snapshot from before the app started is still in `backups/` either way.

### The web UI from the terminal

| Key | Does |
| --- | --- |
| `w` | Start the web UI and open your browser |
| `w` again | Stop it |

While it's running, the status line shows the address.

---

## The web UI

Richer editing, at <http://localhost:4321>. Same notes, live at the same time as
the terminal — edit in one, press `r` in the terminal (or refresh the browser)
to see it.

**Toolbar:** heading, bold, italic, strikethrough, blockquote, inline code, code
block, link, bulleted list, numbered list, task list, horizontal rule. Every
button works on your selection, and pressing it again toggles the formatting off.

**Shortcuts:** `Ctrl+B` bold, `Ctrl+I` italic, `Ctrl+K` link, `Ctrl+S` save.

**Write / Preview tabs** while editing, and a clean read-only view after saving
with an Edit button to go back. There's a search box, a light/dark toggle that
follows your system by default, and deleting asks for confirmation.

**The trash**, the same one the terminal shows. The bin at the foot of the
sidebar carries a count of what is recoverable and opens it: Restore on each
note, a permanent delete per note, and Empty trash. Deleting raises a message
with an Undo button, and an undo arrow sits beside the bin for as long as there
is something to undo. Both permanent actions ask first.

Everything the terminal can do, the browser can do too, and the other way
around — the only exceptions are the ones that only make sense in one place:
`$EDITOR` hand-off and `o` to open a link belong to the terminal, since a
browser already clicks links itself.

---

## Writing notes

GitHub-flavoured Markdown:

~~~markdown
# A heading

**bold**, *italic*, ~~strikethrough~~, `inline code`

- a bullet
- another
   - nested under it

1. numbered
2. lists

- [ ] a task
- [x] a finished task

> a quote

[link text](https://example.com)

| a | table |
| - | ----- |
| 1 | 2     |
~~~

Fenced code blocks work too, written with three backticks or three tildes.

A single newline is a line break, as in GitHub Gists — you don't need two.
Both the terminal and the browser treat it the same way.

**Tables need their separator row.** A header row alone isn't a table; the
`|---|` line under it is what makes one:

~~~markdown
| col1 | col2 | col3 |
| ---- | ---- | ---- |
| 10   | 20   | 30   |
~~~

**Task lists carry a bullet and a box**, `• [ ]` open and `• [✓]` done, the
check in green. This departs from GitHub, which hides the bullet and shows only
the checkbox. The brackets are ASCII because common monospace fonts do not
carry `☐`/`☑`, and the substituted glyph is drawn wider, which swallows the
space after the box.

**Blank lines between groups of bullets are kept.** Markdown normally collapses
them, so the app restores the gap in the preview. Runs come out as 1, 3, 5, 7…
lines: an odd number of blank lines is exact, an even number lands one short.
Numbered lists are left alone, because splitting one would restart it at 1.

---

## Your notes on disk

```
~/.local/share/notes/notes.db          your notes, one SQLite file
~/.local/share/notes/backups/          automatic snapshots
```

Back everything up by copying that one file.

**Snapshots happen on their own.** Every time `note` starts, it copies the
database into `backups/` and keeps the most recent **10**. To go back to one:

```bash
cp ~/.local/share/notes/backups/notes-20260913-140331.db \
   ~/.local/share/notes/notes.db
```

**Try things safely** on a throwaway database, without touching your real notes:

```bash
note --db /tmp/scratch.db
```

Nothing leaves your machine. The web UI binds to `127.0.0.1`, so it isn't
reachable from anywhere else on your network.

---

## The HTTP API

Available whenever the web UI is running, for scripting against your notes.

| Method | Path | Does |
| --- | --- | --- |
| `GET` | `/api/health` | Check it's up |
| `GET` | `/api/notes` | List every note, newest first |
| `GET` | `/api/notes?q=term` | Search titles and bodies |
| `GET` | `/api/notes/:id` | Fetch one note |
| `POST` | `/api/notes` | Create — `{"title": "...", "content": "..."}` |
| `PUT` | `/api/notes/:id` | Update — same shape |
| `DELETE` | `/api/notes/:id` | Move to the trash |
| `GET` | `/api/trash` | List what is recoverable |
| `POST` | `/api/trash/:id` | Restore a note |
| `DELETE` | `/api/trash/:id` | Delete a note for good |
| `DELETE` | `/api/trash` | Empty the trash |

`title` may be omitted or empty; it's derived from the first line.

The trash routes only act on trashed notes: a live note returns 404, so nothing
can be destroyed without being trashed first.

```bash
note --web &
curl localhost:4321/api/notes
curl -X POST localhost:4321/api/notes \
  -H 'Content-Type: application/json' \
  -d '{"content": "# From the shell\n\n- [ ] it works"}'
```

---

## Working on the code

```bash
./build.sh          # build the web bundle into the binary, compile ./note
go test ./...       # store and terminal UI behaviour
go vet ./...
```

Developing the web UI with hot reload, against a running API:

```bash
note --web &
npm --prefix client run dev     # http://localhost:5173
```

| Path | What lives there |
| --- | --- |
| `main.go` | Entry point and flags |
| `internal/store/` | SQLite: the single source of truth |
| `internal/tui/` | The terminal interface |
| `internal/web/` | HTTP API and the embedded web bundle |
| `client/` | React 18 + Vite + Tailwind + shadcn/ui |
