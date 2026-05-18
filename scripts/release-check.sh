#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SKIP_REBUILD=0

for arg in "$@"; do
  case "$arg" in
    --skip-rebuild-local-app)
      SKIP_REBUILD=1
      ;;
    *)
      echo "Unknown argument: $arg" >&2
      exit 1
      ;;
  esac
done

cd "$ROOT_DIR"

./scripts/sync-version.sh
./scripts/check-version-sync.sh
./scripts/check-release-notes.sh
npm run build --prefix desktop/frontend
go test ./...
go build -o lingma-ipc-proxy ./cmd/lingma-ipc-proxy

if [[ "$SKIP_REBUILD" -eq 0 ]]; then
  if [[ "$(uname -s)" == "Darwin" ]]; then
    ./scripts/rebuild-local-app.sh
  else
    echo "Skipping ./scripts/rebuild-local-app.sh because this host is not macOS."
  fi
else
  echo "Skipping ./scripts/rebuild-local-app.sh by request."
fi

echo "Release readiness check passed."
