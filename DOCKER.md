# Docker workflow for this project

Everything below is run from the project root, where the compose files live.
Compose only works from a directory containing one (or with `-f path/to/file`).

**There are two stacks:**

| File | Used by | What it runs |
| --- | --- | --- |
| `docker-compose.yml` | `./notes` | The real app: nginx serving a built bundle, on `127.0.0.1:8080` |
| `docker-compose.dev.yml` | `./notes dev` | Hot-reload dev: Vite on `:5173`, API on `:4000`, Postgres on `:5432` |

The default stack is what a user runs. The dev stack is what you hack on, and
it's the one most examples below use — add `-f docker-compose.dev.yml` to target
it explicitly:

```bash
docker compose -f docker-compose.dev.yml up -d
```

If you just want the app, you never need any of this: `./notes` wraps it.

---

## The mental model

Four concepts, in the order they turn into each other:

| Concept | What it is | Analogy |
| --- | --- | --- |
| **Dockerfile** | A recipe: "start from node:20, copy package.json, run npm install" | The recipe |
| **Image** | The built, frozen result of that recipe. Read-only. | The cake |
| **Container** | A running instance of an image. Disposable. | The slice being eaten |
| **Volume** | Storage that outlives the container | The fridge |

`docker-compose.yml` describes **services** — here `db`, `server`, `client` —
and wires them into one private network where they reach each other **by
service name** (`server` connects to `db:5432`, not `localhost:5432`; nginx
proxies to `server:4000`).

Two networking rules trip everyone up once:

- **Inside** the compose network, use service names: `postgres://...@db:5432/notes`.
- **From your host** (browser, psql, curl), use `localhost:<published port>`,
  and only ports listed under `ports:` are reachable at all.

`ports: "5173:5173"` means *host port 5173 → container port 5173*. Left of the
colon is always yours.

---

## The everyday flow

### 1. Start everything

```bash
docker compose up
```

This one command: builds any image that doesn't exist yet → creates the network
and volumes → starts `db`, waits for its healthcheck to pass → starts `server` →
starts `client`. Logs from all three stream into your terminal, colour-coded.
`Ctrl+C` stops them.

Add `-d` (detached) to hand the terminal back:

```bash
docker compose up -d
```

That's what's running now. The `depends_on: condition: service_healthy` in the
compose file is why the API never starts against a Postgres that isn't ready.

### 2. See what's running

```bash
docker compose ps          # this project's containers + their published ports
docker ps                  # every container on the machine
```

### 3. Watch the logs

```bash
docker compose logs -f            # all services, follow
docker compose logs -f server     # just the API
docker compose logs --tail 50 db  # last 50 lines
```

This is your primary debugging tool. A container that keeps restarting is
almost always explained by its last 20 log lines.

### 4. Edit code

Just edit files on your host. `server/` and `client/` are bind-mounted into the
containers, so changes appear inside instantly — Vite hot-reloads the browser,
`node --watch` restarts the API. **No Docker command needed.**

### 5. Stop

```bash
docker compose stop    # stop containers, keep them and the data
docker compose down    # stop AND delete containers + network (data survives)
docker compose down -v # ...and delete the volumes: database wiped
```

`down -v` is the destructive one. It's also how you re-run `db/init.sql`, since
Postgres only executes that on a **completely empty** data directory.

Note the two stacks store data differently: the dev stack uses a named Docker
volume (`pgdata`), while the production stack bind-mounts `./data/postgres` so
your notes sit in a folder you can see. `docker compose down -v` wipes the dev
volume; `./notes reset` wipes the production folder. They're separate databases,
so notes written in dev won't appear in the real app.

---

## When do I need `--build`?

The single most common confusion. Rule of thumb:

| You changed | Command |
| --- | --- |
| Any file in `server/src` or `client/src` | nothing — hot reload handles it |
| `docker-compose.yml` (ports, env vars) | `docker compose up -d` |
| `package.json` (added a dependency) | `docker compose up -d --build <service> --renew-anon-volumes` |
| A `Dockerfile` | `docker compose up -d --build <service>` |
| `db/init.sql` | `./notes reset` (prod) or `down -v && up -d` (dev) |

Why `--renew-anon-volumes` for dependencies? `node_modules` lives in an
anonymous volume (see below). Rebuilding the image installs the new package into
the *image*, but the old volume would keep shadowing it. That flag throws the
stale volume away. This is exactly what I ran after adding Tailwind and Radix.

---

## The node_modules trick

In the compose file each Node service has two volume lines:

```yaml
volumes:
  - ./client:/app        # your source, live-mounted
  - /app/node_modules    # anonymous volume, no host path
```

The first line makes `/app` inside the container *become* your host folder —
which would hide the `node_modules` that `npm install` put there during the
build. The second line re-mounts `/app/node_modules` as container-only storage,
so the container keeps the Linux-native packages it installed while your source
stays live.

Side effect: you'll see an **empty `client/node_modules` folder on your host**.
That's just the mountpoint Docker created. Ignore it; it's gitignored.

---

## Running commands inside a container

`exec` runs a command in an **already-running** container:

```bash
docker compose exec server sh              # shell inside the API container
docker compose exec client npm run build   # production build
docker compose exec db psql -U notes -d notes
```

Handy psql once you're in: `\dt` lists tables, `\d notes` describes one,
`SELECT id, title FROM notes;`, `\q` quits.

`run` starts a **throwaway** container instead — use it when the service isn't
up, and add `--rm` so it cleans itself up:

```bash
docker compose run --rm server npm install some-package
```

⚠️ A file created by a container lands on your host owned by **root** (the
container runs as root). If `rm` gives you "Permission denied", delete it from
inside the container instead:

```bash
docker compose exec client rm -rf /app/dist
```

---

## Testing / verifying a change

The loop I used while building this:

```bash
# 1. is everything up and healthy?
docker compose ps

# 2. is the API alive?
curl localhost:4000/api/health

# 3. does a real write work end to end?
curl -X POST localhost:4000/api/notes \
  -H 'Content-Type: application/json' \
  -d '{"content":"# test"}'

# 4. did it actually land in Postgres?
docker compose exec db psql -U notes -d notes -c 'SELECT id, title FROM notes;'

# 5. does the frontend compile and proxy to the API?
curl -s -o /dev/null -w '%{http_code}\n' localhost:5173
curl localhost:5173/api/health
docker compose exec client npm run build   # catches every import/JSX error

# 6. clean up the test row
curl -X DELETE localhost:4000/api/notes/<id>
```

Step 5 is worth knowing: hitting `localhost:5173/api/health` proves the Vite dev
proxy is forwarding to `server:4000` over the internal network — the exact path
the browser uses.

---

## When something is broken

```bash
docker compose ps                      # is it even running? restarting?
docker compose logs --tail 50 server   # why did it die?
docker compose exec server sh          # poke around inside
docker compose restart server          # kick one service
```

Common causes:

- **`port is already allocated`** — something else owns that host port. Find it
  with `ss -tulpn | grep 5173`, or change the port in `.env`.
- **API can't reach the DB** — check it's using `db` as the host, not `localhost`.
- **Code changes not appearing** — the bind mount or the watcher. Check the
  `volumes:` path, then `docker compose restart <service>`.
- **Container restart-loops** — read the last log lines; it's usually a crash on
  startup, not Docker.

---

## Housekeeping

```bash
docker compose images        # images this project uses
docker system df             # how much disk Docker is eating
docker system prune          # delete stopped containers + dangling images
docker system prune -a --volumes   # aggressive: also unused images AND volumes
```

`prune -a --volumes` will delete your database volume if the stack is down.

---

## Full reset

When you want to start from absolute zero:

```bash
docker compose down -v          # containers, network, volumes, database
docker compose up --build       # rebuild images from scratch, reseed the DB
```
