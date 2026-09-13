#!/usr/bin/env bash
# Notes - a local-first note taking app.
# Usage: ./notes [start|stop|logs|backup|restore|update|dev|reset|status]
set -euo pipefail

cd "$(dirname "$0")"

PORT="${NOTES_PORT:-8080}"
URL="http://localhost:${PORT}"
DEV_URL="http://localhost:5173"

c_ok()   { printf '\033[32m%s\033[0m\n' "$*"; }
c_info() { printf '\033[36m%s\033[0m\n' "$*"; }
c_warn() { printf '\033[33m%s\033[0m\n' "$*"; }
c_err()  { printf '\033[31m%s\033[0m\n' "$*" >&2; }

require_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    c_err "Docker isn't installed."
    c_err "Get it here: https://docs.docker.com/get-docker/"
    exit 1
  fi
  if ! docker info >/dev/null 2>&1; then
    c_err "Docker is installed but not running. Start Docker Desktop (or 'sudo systemctl start docker') and try again."
    exit 1
  fi
  if ! docker compose version >/dev/null 2>&1; then
    c_err "This needs Docker Compose v2 ('docker compose'). Please update Docker."
    exit 1
  fi
}

open_browser() {
  local url="$1"
  if command -v xdg-open >/dev/null 2>&1; then xdg-open "$url" >/dev/null 2>&1 &
  elif command -v open >/dev/null 2>&1; then open "$url" >/dev/null 2>&1 &
  fi
}

wait_for() {
  local url="$1" tries="${2:-60}"
  for _ in $(seq 1 "$tries"); do
    if curl -sf -o /dev/null "$url"; then return 0; fi
    sleep 1
  done
  return 1
}

cmd_start() {
  require_docker
  c_info "Starting Notes… (first run downloads and builds images, give it a few minutes)"
  docker compose up -d --build
  if wait_for "$URL"; then
    c_ok "Notes is running at ${URL}"
    echo "Your data is stored in ./data - it never leaves this machine."
    echo "Stop it any time with: ./notes stop"
    open_browser "$URL"
  else
    c_err "It didn't come up in time. Check the logs with: ./notes logs"
    exit 1
  fi
}

cmd_stop() {
  require_docker
  docker compose down
  c_ok "Notes stopped. Your notes are safe in ./data"
}

cmd_status() {
  require_docker
  docker compose ps
}

cmd_logs() {
  require_docker
  docker compose logs -f "${1:-}"
}

cmd_backup() {
  require_docker
  mkdir -p backups
  local out="backups/notes-$(date +%Y%m%d-%H%M%S).sql"
  docker compose exec -T db pg_dump -U notes -d notes > "$out"
  c_ok "Backed up to $out"
}

cmd_restore() {
  require_docker
  local file="${1:-}"
  if [ -z "$file" ] || [ ! -f "$file" ]; then
    c_err "Usage: ./notes restore backups/notes-YYYYMMDD-HHMMSS.sql"
    exit 1
  fi
  c_warn "This replaces everything currently in the database with $file"
  read -r -p "Type 'yes' to continue: " reply
  [ "$reply" = "yes" ] || { echo "Cancelled."; exit 0; }
  docker compose exec -T db psql -U notes -d notes -c \
    'DROP SCHEMA public CASCADE; CREATE SCHEMA public;' >/dev/null
  docker compose exec -T db psql -U notes -d notes < "$file" >/dev/null
  c_ok "Restored from $file"
}

cmd_update() {
  require_docker
  if [ -d .git ]; then git pull --ff-only; fi
  docker compose up -d --build
  c_ok "Updated. Notes is running at ${URL}"
}

cmd_dev() {
  require_docker
  c_info "Starting the development stack (hot reload)…"
  docker compose -f docker-compose.dev.yml up -d --build
  if wait_for "$DEV_URL"; then
    c_ok "Dev server running at ${DEV_URL} (API on :4000, Postgres on :5432)"
    echo "Stop it with: docker compose -f docker-compose.dev.yml down"
    open_browser "$DEV_URL"
  else
    c_err "Dev stack didn't come up. Logs: docker compose -f docker-compose.dev.yml logs"
    exit 1
  fi
}

cmd_reset() {
  require_docker
  c_warn "This permanently deletes ALL your notes and starts over."
  read -r -p "Type 'delete' to confirm: " reply
  [ "$reply" = "delete" ] || { echo "Cancelled."; exit 0; }
  docker compose down
  # The data dir is owned by the postgres user inside the container,
  # so remove it from inside a container rather than with host permissions.
  docker run --rm -v "$PWD/data:/data" alpine sh -c 'rm -rf /data/postgres'
  c_ok "Wiped. Run ./notes to start fresh."
}

case "${1:-start}" in
  start|up|"")  cmd_start ;;
  stop|down)    cmd_stop ;;
  status|ps)    cmd_status ;;
  logs)         shift; cmd_logs "${1:-}" ;;
  backup)       cmd_backup ;;
  restore)      shift; cmd_restore "${1:-}" ;;
  update)       cmd_update ;;
  dev)          cmd_dev ;;
  reset)        cmd_reset ;;
  -h|--help|help)
    cat <<'HELP'
Notes - local-first note taking

  ./notes            Start the app and open it in your browser
  ./notes stop       Stop the app (your notes stay on disk)
  ./notes status     Show what's running
  ./notes logs       Follow the logs
  ./notes backup     Write a .sql snapshot into ./backups
  ./notes restore F  Restore from a backup file
  ./notes update     Pull the latest version and rebuild
  ./notes dev        Start the development stack with hot reload
  ./notes reset      Delete all notes and start over

Your data lives in ./data and never leaves this machine.
HELP
    ;;
  *) c_err "Unknown command: $1"; echo "Try: ./notes help"; exit 1 ;;
esac
