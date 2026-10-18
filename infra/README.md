# Infrastructure

Only infrastructure that is justified today exists here. Speculative production
infrastructure is deliberately absent.

## Local development

```bash
docker compose -f infra/local/compose.yaml up -d
```

This starts PostgreSQL and Redis with the credentials in `.env.example`. Both use named
volumes, so data survives a restart and is removed with `docker compose down -v`.

```bash
docker compose -f infra/local/compose.yaml down
```

The backend is not containerised for local development. Run it directly, which keeps the
edit and test loop fast:

```bash
cd backend && go run ./cmd/api
```

## Container image

`docker/backend.Dockerfile` builds either backend binary from the repository root:

```bash
docker build -f infra/docker/backend.Dockerfile --build-arg TARGET=api -t nuvro-api .
```

```bash
docker build -f infra/docker/backend.Dockerfile --build-arg TARGET=worker -t nuvro-worker .
```

The image is a static binary on a distroless base running as a non-root user. It contains no
shell and no package manager.

## Deliberately absent

These are future decisions, not oversights. The cloud deployment architecture has not been
selected, so committing configuration for it now would encode a choice nobody has made.

| Concern | Status |
| --- | --- |
| Terraform or other infrastructure as code | Deferred until the cloud target is chosen |
| Kubernetes manifests | Deferred, and possibly never needed for a modular monolith |
| Helm charts | Deferred |
| Managed PostgreSQL and Redis provisioning | Deferred |
| CDN and object storage for travel media | Deferred |
| Secret manager integration | Deferred, configuration is read from the environment today |
| OpenTelemetry collector deployment | Deferred, see `docs/architecture/overview.md` |

When the deployment target is chosen, record it as an architecture decision record under
`docs/adr/` before adding configuration here.
