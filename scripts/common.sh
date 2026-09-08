#!/usr/bin/env sh
set -eu

die() {
  echo "error: $*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

require_value() {
  name="$1"
  eval "value=\${$name:-}"
  [ -n "$value" ] || die "$name is required"
}

job_name() {
  printf 'github-actions-sample-%s\n' "$DEPLOY_ENVIRONMENT"
}

consul_key_base() {
  printf 'railzway/github-actions-sample/%s/deploy\n' "$DEPLOY_ENVIRONMENT"
}
