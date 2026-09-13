# Using note

Everything this app can do, and how to do it.

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

---

## The terminal UI

Two panes: your notes on the left, the selected note rendered on the right. The
bar along the bottom always shows the keys available right now.

### Moving around

| Key | Does |
| --- | --- |
| `j` / `↓` | Next note |
| `k` / `↑` | Previous note |
| `g` | Jump to the first note |
| `G` | Jump to the last note |
| `ctrl+d` | Scroll the preview down half a screen |
| `ctrl+u` | Scroll the preview up half a screen |
| `r` | Reload from disk (picks up changes made in the web UI) |
| `?` | Full help — any key closes it |
| `q` | Quit |

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
| `ctrl+b` | Bold the word under the cursor |
| `alt+i` | Italic the word under the cursor |
| `ctrl+k` | Turn the word into a link, cursor ready for the address |
| `ctrl+e` | Hand the body to `$EDITOR`; save and quit there to come back |
| `↵` | New line (when the body has focus) |

Formatting keys work on the word under the cursor and toggle off if you press
them again. `alt+i` carries italic because **`ctrl+i` is Tab** in every
terminal — they are the same byte — and Tab already switches title and body.

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

**Task lists have no bullet, matching GitHub.** GitHub's own stylesheet sets
`list-style-type: none` on task items and pulls the checkbox into the marker's
place, so `☐` and `☑` stand where a `•` would, and task text lines up with
ordinary bullet text.

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

`title` may be omitted or empty; it's derived from the first line.

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
