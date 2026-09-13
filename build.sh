#!/usr/bin/env bash
# Builds the web UI, embeds it, and compiles the single `notes` binary.
set -euo pipefail
cd "$(dirname "$0")"

export PATH="$HOME/.local/go/bin:$PATH"

if ! command -v go >/dev/null 2>&1; then
  echo "Go is not installed. Get it from https://go.dev/dl" >&2
  exit 1
fi

VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo dev)"

if command -v npm >/dev/null 2>&1; then
  echo "==> Building the web UI"
  npm --prefix client install --no-fund --no-audit --silent
  npm --prefix client run build

  echo "==> Embedding it in the binary"
  rm -rf internal/web/dist
  cp -r client/dist internal/web/dist
else
  echo "==> npm not found; the web UI will show a placeholder"
fi

# go:embed needs the directory to exist even when the UI was never built.
if [ ! -f internal/web/dist/index.html ]; then
  mkdir -p internal/web/dist
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
go build -ldflags "-s -w -X main.version=${VERSION}" -o notes .

echo
echo "Built ./notes ($(du -h notes | cut -f1))"
echo "Run it with:  ./notes"
