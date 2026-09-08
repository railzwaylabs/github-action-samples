#!/usr/bin/env sh
set -eu
. "$(dirname "$0")/common.sh"

require_command consul
require_value DEPLOY_ENVIRONMENT
key_base="$(consul_key_base)"
candidate="$(consul kv get -quiet "$key_base/candidate_image")"
[ -n "$candidate" ] || die "candidate image is missing"
consul kv put "$key_base/last_successful_image" "$candidate" >/dev/null
consul kv put "$key_base/last_successful_at" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" >/dev/null
echo "marked successful: $candidate"
