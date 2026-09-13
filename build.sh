#!/usr/bin/env bash
# Builds the web UI, embeds it, and compiles the single `note` binary.
set -euo pipefail
cd "$(dirname "$0")"

# Go is often installed somewhere the shell's PATH does not cover.
for dir in /usr/local/go/bin "$HOME/.local/go/bin" "$HOME/go/bin" /usr/lib/go/bin; do
  [ -x "$dir/go" ] && PATH="$dir:$PATH"
done
export PATH

if ! command -v go >/dev/null 2>&1; then
  echo "Go is not installed, or not on your PATH. Get it from https://go.dev/dl" >&2
  exit 1
fi

if ! command -v npm >/dev/null 2>&1; then
  echo "Note: npm was not found, so the web UI cannot be built into the binary." >&2
  echo "      The terminal interface will work; install Node to get the web UI." >&2
fi

VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo dev)"

# The one committed file in internal/web/dist. It exists so `go:embed all:dist`
# has something to match on a fresh clone, which is what lets `go build` and
# `go test` work before the web bundle has ever been built. Rewritten verbatim
# after each build so the working tree stays clean.
write_gitkeep() {
  mkdir -p internal/web/dist
  cat > internal/web/dist/.gitkeep <<'KEEP'
The web UI is built into this directory by ./build.sh and is not committed.

This file is committed so that `go:embed all:dist` has something to match,
which means `go build` and `go test` work on a fresh clone without having to
build the web bundle first. Without it the package does not compile at all.
KEEP
}

if command -v npm >/dev/null 2>&1; then
  echo "==> Building the web UI"
  npm --prefix client install --no-fund --no-audit --silent
  npm --prefix client run build

  echo "==> Embedding it in the binary"
  rm -rf internal/web/dist
  cp -r client/dist internal/web/dist
  write_gitkeep
else
  echo "==> npm not found; the web UI will show a placeholder"
fi

# go:embed needs the directory to exist even when the UI was never built.
write_gitkeep
if [ ! -f internal/web/dist/index.html ]; then
  cat > internal/web/dist/index.html <<'HTML'
<!doctype html>
<title>Notes — UI not built</title>
<body style="font:14px system-ui;padding:2rem">
  <h1>The web UI has not been built into this binary.</h1>
  <p>Install Node, then run <code>./build.sh</code> again.</p>
</body>
HTML
fi

echo "==> Compiling"
go build -ldflags "-s -w -X main.version=${VERSION}" -o note .

echo
echo "Built ./note ($(du -h note | cut -f1))"
echo "Run it with:  ./note"
