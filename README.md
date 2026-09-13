# Notes

A note-taking app that runs entirely on your own machine. Markdown editing with
the things you actually want — bold, italics, lists, links, code blocks, live
preview — and **none of your writing ever leaves your computer**.

No account. No cloud. No sync service. No telemetry. Your notes are rows in a
Postgres database in a folder you own, and you can copy, back up or delete that
folder whenever you like.

---

## Setup

You need [Docker](https://docs.docker.com/get-docker/). That's the only
prerequisite — no Node, no Postgres, nothing else to install.

```bash
git clone https://github.com/<you>/notes.git
cd notes
./notes
```

That's it. The first run builds everything (a few minutes), then opens
<http://localhost:8080> in your browser. Every run after that takes seconds.

> **Windows:** run the same commands inside WSL2 or Git Bash.

---

## Using it

| Command | What it does |
| --- | --- |
| `./notes` | Start the app and open it in your browser |
| `./notes stop` | Stop the app (your notes stay on disk) |
| `./notes status` | Show what's running |
| `./notes logs` | Follow the logs |
| `./notes backup` | Write a `.sql` snapshot into `./backups` |
| `./notes restore <file>` | Restore from a backup |
| `./notes update` | Pull the latest version and rebuild |
| `./notes reset` | Delete all notes and start over |
| `./notes dev` | Development stack with hot reload |

Prefer a different port? `NOTES_PORT=9000 ./notes`

---

## Where your data lives

Everything is in **`./data/postgres`**, inside this folder. To move your notes
to another machine, copy the folder — or take a portable snapshot:

```bash
./notes backup        # -> backups/notes-20260913-125456.sql
```

Two things worth knowing:

- The app binds to `127.0.0.1` only, so it isn't reachable from anyone else on
  your network — not even other devices in your house.
- The database and API containers publish **no ports at all**. The only way in
  is the web interface on your own machine.

`./data` is owned by the database user inside the container, so your file
manager may ask for permission to browse it. That's normal — use
`./notes backup` to get a readable copy.

---

## Writing notes

Markdown, GitHub-flavored. Toolbar plus keyboard shortcuts:

| Feature | Shortcut |
| --- | --- |
| Bold | `Ctrl+B` |
| Italic | `Ctrl+I` |
| Link | `Ctrl+K` |
| Save | `Ctrl+S` |

Also in the toolbar: headings, strikethrough, blockquotes, inline code, code
blocks, bulleted / numbered / task lists, and horizontal rules. Every button
works on your selection and toggles off if you press it again.

There's a **Write / Preview** tab pair while editing and a clean **saved view**
afterwards, full-text search across every note, and a light/dark theme that
follows your system by default. A plain newline is a line break, the way it
works in GitHub Gists. Rendered HTML is sanitized before it's displayed.

---

## How it's put together

```
client/   React 18 + Vite + Tailwind v4 + shadcn/ui
server/   Express REST API
db/       Postgres schema and seed
notes     The launcher script
```

Three containers: nginx serves the built frontend and proxies `/api` to the
Express container, which talks to Postgres over a private Docker network.

### Hacking on it

```bash
./notes dev     # http://localhost:5173, hot reload for frontend and API
```

Source folders are bind-mounted, so edits on your machine reload instantly.
See **[DOCKER.md](./DOCKER.md)** for how the Docker side works — what each
command does, when you need `--build`, how to inspect the database, and how to
debug a container that won't start.

shadcn/ui components are copied into `client/src/components/ui/`, so edit them
freely. To add another:

```bash
docker compose -f docker-compose.dev.yml exec client npx shadcn@latest add dialog
```

### API

| Method | Path |
| --- | --- |
| `GET` | `/api/notes?q=` |
| `GET` | `/api/notes/:id` |
| `POST` | `/api/notes` |
| `PUT` | `/api/notes/:id` |
| `DELETE` | `/api/notes/:id` |

---

## License

MIT
