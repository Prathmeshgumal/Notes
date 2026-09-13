# note

A Markdown note taker that lives in your terminal, with a web UI on the same
notes whenever you want one.

Everything stays on your machine — one SQLite file you can copy. No account, no
cloud, no sync service, nothing running in the background.

```
╭ Notes (4) ───────────────╮╭ 14 sep Tasks ────────────────────────────────╮
│                          ││                                              │
│ ▸ 14 sep Tasks           ││  • [✓] Solve 10 DSA questions                │
│   13 sep Daily standup   ││  • [ ] Prepare resume                        │
│   Reading list           ││    • [ ] Review the skills                   │
│   Meeting notes          ││  • [ ] Go for a run                          │
│                          ││                                              │
╰──────────────────────────╯╰──────────────────────────────────────────────╯
 j/k move  ↵ edit  n new  / search  o link  d trash  w web  ? help  q quit
```

## Setup

You need [Go](https://go.dev/dl) and [Node](https://nodejs.org) to build it.
Neither is needed to run it — the result is one self-contained binary.

```bash
git clone https://github.com/Prathmeshgumal/Notes.git
cd Notes
./build.sh
cp note ~/.local/bin/
```

Then, from anywhere:

```bash
note
```

If `note: command not found`, add `~/.local/bin` to your PATH:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc && source ~/.bashrc
```

## Keys

| Key | Does |
| --- | --- |
| `j` `k` | Move up and down |
| `↵` | Edit the selected note |
| `n` | New note |
| `/` | Search titles and bodies |
| `o` | Open a link from the note |
| `d` | Move to the trash |
| `u` | Undo the last delete |
| `T` | Open the trash |
| `w` | Start the web UI |
| `?` | Full help |
| `q` | Quit |

While editing: `ctrl+s` saves, `ctrl+p` previews, `esc` discards, `tab` switches
between title and body, and `ctrl+e` hands the note to your `$EDITOR`. The full
set of formatting keys — bold, italic, strikethrough, code, headings, the three
list types, quotes and rules — is in `?`.

Leave the title blank and the first line becomes it, the way GitHub Gists work.

See **[USAGE.md](./USAGE.md)** for everything — every key, the Markdown rules,
and the HTTP API.

## The web UI

Press `w` in the terminal, or run it on its own:

```bash
note --web          # http://localhost:4321
```

Same notes, same moment — both can be open at once. Everything the terminal can
do the browser can too, including the trash: restore, delete for good and empty.
The browser adds a formatting toolbar, live preview and a light/dark theme. It
binds to `127.0.0.1`, so nothing else on your network can reach it.

## Your notes

```
~/.local/share/notes/notes.db        every note, one file
~/.local/share/notes/backups/        automatic snapshots, the last 10
```

Copy that file to back everything up. Deleting is recoverable: `d` moves a note
to the trash, `u` undoes it, `T` browses the trash, and a snapshot is taken
every time the app starts.

To experiment without touching your real notes:

```bash
note --db /tmp/scratch.db
```

## Built with

| Path | What lives there |
| --- | --- |
| `main.go` | Entry point and flags |
| `internal/store/` | SQLite — the single source of truth |
| `internal/tui/` | The terminal interface, using Bubble Tea |
| `internal/web/` | HTTP API and the embedded web bundle |
| `client/` | React 18, Vite, Tailwind, shadcn/ui |

The React bundle is compiled into the binary, which is why the web UI needs
nothing installed to serve it.

```bash
go test ./...       # store and terminal-UI behaviour
./build.sh          # rebuild the bundle and the binary
```

## License

MIT
