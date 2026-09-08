# GitHub Actions + Nomad sample

A small PostgreSQL-backed order API used to demonstrate CI, staged testing, and immutable Nomad deployments.

## Run locally

```sh
docker compose up --build
curl http://localhost:5173
curl http://localhost:8080/health
curl http://localhost:8080/ready
curl -H 'Content-Type: application/json' \
  -d '{"customer_id":"customer-1","items":[{"product_id":"product-1","quantity":2,"price":1250}]}' \
  http://localhost:8080/orders
```

Endpoints:

- `GET /health` and `GET /ready`
- `GET /orders`
- `POST /orders`
- `GET /orders/{id}`

Orders are persisted in PostgreSQL through GORM. Schema changes are owned by versioned `golang-migrate` SQL files under `db/`; the application server never modifies the schema.

Local Docker Compose enables profiling only on `127.0.0.1:6060`:

```sh
go tool pprof http://localhost:6060/debug/pprof/heap
go tool pprof 'http://localhost:6060/debug/pprof/profile?seconds=30'
curl 'http://localhost:6060/debug/pprof/goroutine?debug=1'
```

Profiling is disabled by default. Enable it with `PROFILING_ENABLED=true`. Keep
`PROFILING_ADDRESS=127.0.0.1:6060` unless access is protected by a private network
or port forwarding.

Apply migrations manually with:

```sh
DATABASE_URL='postgres://postgres:postgres@localhost:5432/orders?sslmode=disable' \
  go run ./cmd/sample migrate up
```

Inspect the current migration version with `go run ./cmd/sample migrate version`. Docker Compose and the Nomad jobspec run `migrate up` automatically before starting the API.

## Frontend development

The React/Vite frontend lives in `apps/sample` and is managed by the pnpm workspace at the repository root. Install dependencies once for every package:

```sh
corepack enable
pnpm install
pnpm dev:web
```

The first root install creates one shared `pnpm-lock.yaml`. Commit that lockfile and use `pnpm install --frozen-lockfile` in CI once it exists.

Open `http://localhost:5173`. The frontend runs separately from the API; Vite proxies API calls to `http://localhost:8080` during local development.

The Playwright suite creates an order through the browser and verifies the persisted result:

```sh
pnpm --filter github-actions-sample-web exec playwright install chromium
PLAYWRIGHT_BASE_URL=http://localhost:8080 pnpm test:e2e
```

## Deployment

CI builds `Dockerfile`, pushes `ghcr.io/<owner>/<repository>:sha-<commit>`, and records its immutable digest in a checksummed release artifact. The release workflow deploys that digest through `deployments/nomad/sample.nomad.hcl`.

See `.github/workflows/README.md` for GitHub environments, variables, runner prerequisites, and the deployment script contract.
