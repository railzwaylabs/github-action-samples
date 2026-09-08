# Pipeline setup

The workflows implement this path:

`CI lanes -> quality gate -> immutable artifact -> staging -> PTR -> QA review -> UAT -> production -> health/smoke -> rollback on failure`

## Required repository setup

1. Set `ENABLE_DEPLOYMENTS=true` after the scripts have been tested against your Nomad and Consul cluster.
2. Create the `staging`, `qa-review`, `uat`, and `production` GitHub environments.
3. Add required reviewers to `qa-review`, `uat`, and `production`. Rejecting an environment is the diagram's fail path back to development.
4. Set `MIN_COVERAGE` to the desired numeric coverage floor. It defaults to `0` while this sample has no tests.
5. Set `STAGING_BASE_URL`; set `HEALTH_URL` on staging and production, and `SMOKE_BASE_URL` on production.
6. Provide a self-hosted runner labeled `self-hosted`, `linux`, and `x64`, with Nomad/Consul credentials in environment secrets.

## Deployment scripts

The release workflow uses these executable repository scripts:

- `scripts/deploy.sh` deploys `RELEASE_ARTIFACT` to `DEPLOY_ENVIRONMENT` and records the prior release.
- `scripts/health-check.sh URL` waits for a bounded period until the service is healthy.
- `scripts/ptr.sh SUITE BASE_URL` runs one PTR suite and can write evidence to `test-results/`.
- `scripts/verify-nomad.sh` waits for a healthy Nomad deployment.
- `scripts/smoke-test.sh BASE_URL` checks critical production behavior.
- `scripts/rollback.sh` restores the last known-good production release.
- `scripts/mark-successful.sh` records the verified digest as the last known-good release.

The Nomad specification is `deployments/nomad/sample.nomad.hcl`. `deploy.sh` supplies its environment, immutable image digest, datacenters, replica count, and optional Traefik host.

Configure `NOMAD_ADDR`, `NOMAD_TOKEN`, `CONSUL_HTTP_ADDR`, `CONSUL_HTTP_TOKEN`, and `DATABASE_URL` as environment secrets. Optional environment variables are `NOMAD_DATACENTERS` (HCL list, default `["dc1"]`), `NOMAD_COUNT` (default `1`), and `APP_HOST` for Traefik routing. Nomad clients also need credentials to pull the private GHCR image, or the package must be public.

Branch protection for `main` should require the `CI quality gate` check.

For a manual release rerun, copy both the CI run ID and its full commit SHA from the successful CI run. The release workflow requires both so an artifact cannot be promoted under a different commit identity.
