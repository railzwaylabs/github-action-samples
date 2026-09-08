#!/usr/bin/env sh
set -eu
. "$(dirname "$0")/common.sh"

require_command curl
require_command jq
base_url="${1:-}"
[ -n "$base_url" ] || die "usage: smoke-test.sh BASE_URL"
base_url="${base_url%/}"

response="$(curl --fail --silent --show-error --max-time 10 \
  -H 'Content-Type: application/json' \
  -d '{"customer_id":"deployment-smoke","items":[{"product_id":"sample","quantity":2,"price":1250}]}' \
  "$base_url/orders")"
order_id="$(printf '%s' "$response" | jq -er '.id')"
subtotal="$(printf '%s' "$response" | jq -er '.items[0].subtotal')"
[ "$subtotal" = 2500 ] || die "unexpected subtotal: $subtotal"
curl --fail --silent --show-error --max-time 10 "$base_url/orders/$order_id" | jq -e --arg id "$order_id" '.id == $id' >/dev/null
echo "production smoke test passed for order $order_id"
