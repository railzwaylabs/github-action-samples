#!/usr/bin/env sh
set -eu
. "$(dirname "$0")/common.sh"

require_command nomad
require_command consul
require_value DEPLOY_ENVIRONMENT
key_base="$(consul_key_base)"
previous="$(consul kv get -quiet "$key_base/previous_image" 2>/dev/null || true)"
[ -n "$previous" ] || die "no previous image recorded at $key_base/previous_image"

nomad job run -detach \
  -var="environment=$DEPLOY_ENVIRONMENT" \
  -var="image_ref=$previous" \
  -var="datacenters=${NOMAD_DATACENTERS:-[\"dc1\"]}" \
  -var="count=${NOMAD_COUNT:-1}" \
  -var="host=${APP_HOST:-}" \
  deployments/nomad/sample.nomad.hcl
consul kv put "$key_base/rollback_image" "$previous" >/dev/null
echo "rollback submitted with $previous"
