# End-to-end test strategy

The E2E suites are black-box checks. They call only externally visible HTTP or browser interfaces and use a real PostgreSQL service in CI.

CI applies the same versioned migrations used by Nomad before either E2E suite starts. A migration failure blocks test execution and artifact promotion.

| Layer | Scenario | Main risk covered | Evidence |
| --- | --- | --- | --- |
| Backend | `/health` and `/ready` | unhealthy release promoted | Go test JSON |
| Backend | create, get, and list order | persistence or API contract regression | Go test JSON |
| Backend | two-item subtotal `2500 + 1499` | financial calculation drift | Go assertion + JSON |
| Backend | missing customer/items, zero quantity, negative price | invalid financial input accepted | Go assertion + JSON |
| Backend | unknown order | incorrect not-found contract | Go assertion + JSON |
| Frontend | create and reload order | broken UI-to-API persistence flow | Playwright HTML, JUnit, trace on retry |
| Frontend | invalid quantity | missing client constraint | Playwright HTML and JUnit |
| Frontend | simulated API 503 | silent UI failure or lost user input | Playwright HTML and JUnit |

## Local execution

Start the complete application:

```sh
docker compose up --build
```

Run backend E2E:

```sh
./scripts/backend-e2e.sh http://localhost:8080
```

Run frontend E2E:

```sh
corepack enable
pnpm install
pnpm --filter github-actions-sample-web exec playwright install chromium
PLAYWRIGHT_BASE_URL=http://localhost:5173 pnpm test:e2e
```

## Staging rollback drill

A rollback test requires two different immutable image digests. Treat version A
as the current known-good release, then deploy version B as the candidate:

1. Deploy version A, complete health and smoke checks, and run
   `scripts/mark-successful.sh`.
2. Deploy a different version B. The deploy script records A as
   `previous_image` before submitting B.
3. Set `DEPLOY_ENVIRONMENT=staging` and run `scripts/rollback.sh`.
4. Run `scripts/verify-nomad.sh`, `scripts/health-check.sh`, and
   `scripts/smoke-test.sh` against staging.
5. Confirm the running Nomad task image equals the Consul
   `last_successful_image` value for version A.

Do not test rollback for the first time in production. Reusing the same digest
for A and B is not a rollback test because no distinct `previous_image` is
recorded. The drill must also confirm that migrations deployed by B remain
compatible with application version A.

Test records use unique customer IDs and contain no personal data. The current API has no deletion endpoint, so persistent shared environments require scheduled test-data cleanup by the `backend-e2e-*` and `e2e-customer-*` prefixes.
