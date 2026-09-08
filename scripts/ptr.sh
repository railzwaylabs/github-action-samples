#!/usr/bin/env sh
set -eu
. "$(dirname "$0")/common.sh"

require_command curl
require_command jq
suite="${1:-}"
base_url="${2:-}"
[ -n "$suite" ] && [ -n "$base_url" ] || die "usage: ptr.sh SUITE BASE_URL"
base_url="${base_url%/}"
mkdir -p test-results

case "$suite" in
  frontend-e2e)
    require_command pnpm
    PLAYWRIGHT_BASE_URL="$base_url" pnpm --filter github-actions-sample-web test:e2e
    cp -R apps/sample/playwright-report "test-results/frontend-playwright-report"
    cp apps/sample/test-results/results.xml "test-results/frontend-e2e.xml"
    printf '{"suite":"frontend-e2e","status":"passed"}\n' > "test-results/$suite.json"
    ;;
  public-api)
    ./scripts/backend-e2e.sh "$base_url"
    ;;
  backend-critical-flow|cross-service-business-flow)
    response="$(curl --fail --silent --show-error -H 'Content-Type: application/json' \
      -d "{\"customer_id\":\"ptr-$suite\",\"items\":[{\"product_id\":\"sample\",\"quantity\":2,\"price\":1250}]}" \
      "$base_url/orders")"
    printf '%s\n' "$response" > "test-results/$suite.json"
    order_id="$(printf '%s' "$response" | jq -er '.id')"
    curl --fail --silent --show-error "$base_url/orders/$order_id" | jq -e '.items[0].subtotal == 2500' >/dev/null
    ;;
  *) die "unknown PTR suite: $suite" ;;
esac
echo "PTR suite passed: $suite"
