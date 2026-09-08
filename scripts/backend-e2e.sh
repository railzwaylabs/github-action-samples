#!/usr/bin/env sh
set -eu
. "$(dirname "$0")/common.sh"

require_command go
base_url="${1:-${E2E_BASE_URL:-http://localhost:8080}}"
results_dir="${E2E_RESULTS_DIR:-test-results}"
mkdir -p "$results_dir"

echo "running backend E2E against $base_url"
set +e
E2E_BASE_URL="$base_url" go test -tags=e2e -count=1 -timeout=2m -json ./tests/e2e/backend > "$results_dir/backend-e2e.json"
status=$?
set -e
cat "$results_dir/backend-e2e.json"
exit "$status"
