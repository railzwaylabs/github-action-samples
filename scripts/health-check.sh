#!/usr/bin/env sh
set -eu
. "$(dirname "$0")/common.sh"

require_command curl
url="${1:-}"
[ -n "$url" ] || die "usage: health-check.sh URL"

i=1
while [ "$i" -le "${HEALTH_ATTEMPTS:-30}" ]; do
  if curl --fail --silent --show-error --max-time 5 "$url" >/dev/null; then
    echo "healthy: $url"
    exit 0
  fi
  echo "waiting for health endpoint ($i/${HEALTH_ATTEMPTS:-30})"
  sleep "${HEALTH_INTERVAL_SECONDS:-10}"
  i=$((i + 1))
done
die "health check failed: $url"
