#!/usr/bin/env sh
set -eu
. "$(dirname "$0")/common.sh"

require_command nomad
require_command consul
require_command tar
require_value DEPLOY_ENVIRONMENT
require_value RELEASE_ARTIFACT
require_value DATABASE_URL

release_metadata="$(tar -xOf "$RELEASE_ARTIFACT" release.env)" || die "invalid release artifact"
eval "$release_metadata"
require_value IMAGE_REF
require_value RELEASE_SHA
case "$IMAGE_REF" in
  *@sha256:*) ;;
  *) die "IMAGE_REF must be pinned to an sha256 digest" ;;
esac

key_base="$(consul_key_base)"
consul kv put "railzway/github-actions-sample/$DEPLOY_ENVIRONMENT/database_url" "$DATABASE_URL" >/dev/null
current="$(consul kv get -quiet "$key_base/last_successful_image" 2>/dev/null || true)"
if [ -n "$current" ] && [ "$current" != "$IMAGE_REF" ]; then
  consul kv put "$key_base/previous_image" "$current" >/dev/null
fi
consul kv put "$key_base/candidate_image" "$IMAGE_REF" >/dev/null
consul kv put "$key_base/candidate_sha" "$RELEASE_SHA" >/dev/null

nomad job run -detach \
  -var="environment=$DEPLOY_ENVIRONMENT" \
  -var="image_ref=$IMAGE_REF" \
  -var="datacenters=${NOMAD_DATACENTERS:-[\"dc1\"]}" \
  -var="count=${NOMAD_COUNT:-1}" \
  -var="host=${APP_HOST:-}" \
  deployments/nomad/sample.nomad.hcl

echo "submitted $(job_name) with $IMAGE_REF"
