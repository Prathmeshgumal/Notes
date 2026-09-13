# Notes

A Markdown note taker for people who live in a terminal. Keyboard-driven TUI in
the spirit of lazygit, with the same notes available in a proper web UI whenever
you want one — from the same single binary.

Everything stays on your machine. No account, no cloud, no sync service.

```
╭ Notes (4) ───────────────╮╭ Welcome to your notes ────────────────────────╮
│                          ││                                               │
│ ▸ Welcome to your notes   ││   # Welcome                                  │
│   Jobs                   ││                                               │
│   Reading list           ││   This editor speaks Markdown, just like      │
│   Meeting 12 Sep         ││   GitHub Gists.                               │
│                          ││                                               │
│                          ││   • Bold, italics, links, code                │
│                          ││   • Task lists and tables                     │
│                          ││                                               │
╰──────────────────────────╯╰───────────────────────────────────────────────╯
 j/k move   ↵ edit   n new   / search   d delete   w web   ? help   q quit
```

## Install

Needs [Go](https://go.dev/dl) to build, and Node only if you want to rebuild the
web UI. The result is one self-contained binary.

```bash
git clone https://github.com/<you>/note.git
cd note
./build.sh
cp note ~/.local/bin/      # anywhere on your PATH
```

Then, from anywhere:

```bash
note
```

See **[USAGE.md](./USAGE.md)** for the full reference.

## Keys

| Key | Does |
| --- | --- |
| `j` / `k` | move down / up (arrows work too) |
| `g` / `G` | jump to first / last |
| `↵` | edit the selected note |
| `n` | new note |
| `/` | search titles and bodies, `esc` clears |
| `o` | open a link from the note (again for the next) |
| `d` | move to trash, asks first |
| `u` | undo the last delete |
| `w` | start the web UI and open a browser |
| `r` | reload from disk |
| `?` | full help |
| `q` | quit |

While editing: `ctrl+s` saves, `tab` switches between title and body, `esc`
discards, and `ctrl+e` hands the note to `$EDITOR` (falling back to nvim, vim,
nano or vi) so you can write in your own setup and come back.

Titles are optional — leave one blank and the first line of the note becomes the
title, the way GitHub Gists do it.

## The web UI

Press `w` in the TUI, or run it on its own:

```bash
note --web            # http://localhost:4321
note --web --port 9000
```

Same notes, same database, live at the same time — SQLite's WAL mode means the
terminal and the browser can both be open without stepping on each other. The
React bundle is compiled into the binary, so there's nothing else to serve or
deploy. It binds to `127.0.0.1`, so nothing on your network can reach it.

The web side has the richer editor: a formatting toolbar, live preview, and
light/dark theming.

## Your data

One SQLite file:

```
~/.local/share/notes/notes.db
```

Back it up by copying that file.

### Nothing is deleted in a hurry

`d` moves a note to a trash rather than destroying it. `u` brings back the last
one you trashed, and trashed notes stay recoverable for 30 days before being
purged.

Every time `note` starts it also copies the database to
`~/.local/share/notes/backups/`, keeping the last 10. To go back to one:

```bash
cp ~/.local/share/notes/backups/notes-20260913-140331.db \
   ~/.local/share/notes/notes.db
```

### Trying things out safely

Point it at a throwaway database to experiment without touching your notes:

```bash
note --db /tmp/scratch.db
``` Point somewhere else with `--db /path/to.db` or
`NOTES_DB=/path/to.db`. No server, no daemon, nothing running when you're not
using it.

## Development

```
main.go              entry point and flags
internal/store/      SQLite: the single source of truth
internal/tui/        Bubble Tea terminal interface
internal/web/        HTTP API + the embedded React bundle
client/              React 18 + Vite + Tailwind + shadcn/ui
```

```bash
go test ./...        # store and TUI behaviour
./build.sh           # rebuild the bundle and the binary
```

The web UI can be developed with hot reload against a running API:

```bash
note --web &
npm --prefix client run dev     # http://localhost:5173
```

An older Docker + Postgres deployment lives in `docker-compose.yml` and
`scripts/docker-stack.sh` — useful if you ever want to host this for more than
one person. See [DOCKER.md](./DOCKER.md). The single binary above supersedes it
for personal use.

## License

MIT
