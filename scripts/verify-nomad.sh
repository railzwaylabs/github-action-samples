#!/usr/bin/env sh
set -eu
. "$(dirname "$0")/common.sh"

require_command nomad
require_command jq
require_value DEPLOY_ENVIRONMENT
job="$(job_name)"
attempts="${VERIFY_ATTEMPTS:-30}"

i=1
while [ "$i" -le "$attempts" ]; do
  running="$(nomad job allocs -json "$job" | jq '[.[] | select(.ClientStatus == "running")] | length')"
  desired="$(nomad job status -json "$job" | jq '[.TaskGroups[].Count] | add // 0')"
  if [ "$desired" -gt 0 ] && [ "$running" -ge "$desired" ]; then
    echo "$job has $running/$desired running allocations"
    exit 0
  fi
  echo "waiting for $job allocations ($i/$attempts)"
  sleep "${VERIFY_INTERVAL_SECONDS:-10}"
  i=$((i + 1))
done
die "$job did not become ready"
