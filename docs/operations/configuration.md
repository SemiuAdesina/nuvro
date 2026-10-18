# Configuration reference

Every setting is read from the environment at startup and validated before the process serves
traffic. An invalid configuration fails the process rather than starting a service that is
subtly wrong. `NUVRO_ENV` decides which settings are required: `development` allows almost
everything to be absent, `staging` and `production` do not.

## Process

| Variable | Default | Notes |
| --- | --- | --- |
| `NUVRO_ENV` | `development` | `development`, `staging` or `production`. |
| `NUVRO_LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error`. |
| `NUVRO_HTTP_PORT` | `8080` | |
| `NUVRO_HTTP_READ_HEADER_TIMEOUT` | `5s` | |
| `NUVRO_HTTP_READ_TIMEOUT` | `15s` | |
| `NUVRO_HTTP_WRITE_TIMEOUT` | `30s` | |
| `NUVRO_HTTP_IDLE_TIMEOUT` | `120s` | |
| `NUVRO_HTTP_SHUTDOWN_TIMEOUT` | `20s` | In-flight requests finish before the process exits. |
| `NUVRO_HTTP_MAX_HEADER_BYTES` | `1048576` | |

## Backing services

| Variable | Default | Notes |
| --- | --- | --- |
| `NUVRO_DATABASE_URL` | empty | Required outside development. `postgres` or `postgresql` scheme. |
| `NUVRO_REDIS_URL` | empty | Required outside development. `redis` or `rediss` scheme. |

Without a database URL the process starts with no product handlers at all. That is a
development convenience, not a degraded production mode.

## Identity

| Variable | Default | Notes |
| --- | --- | --- |
| `NUVRO_APPLE_AUDIENCE` | empty | The bundle identifier. Required outside development. |
| `NUVRO_APPLE_ISSUER` | `https://appleid.apple.com` | |
| `NUVRO_APPLE_KEY_ENDPOINT` | `https://appleid.apple.com/auth/keys` | Must be https. |
| `NUVRO_APPLE_KEY_TTL` | `1h` | How long a fetched signing key is trusted before refetching. |
| `NUVRO_APPLE_TEAM_ID` | empty | Required outside development, for token revocation. |
| `NUVRO_APPLE_KEY_ID` | empty | The `kid` of the Sign in with Apple key. |
| `NUVRO_APPLE_PRIVATE_KEY` | empty | Base64 of the PKCS#8 PEM `.p8`. Validated at startup. |
| `NUVRO_APPLE_TOKEN_ENDPOINT` | `https://appleid.apple.com/auth/token` | Must be https. |
| `NUVRO_APPLE_REVOKE_ENDPOINT` | `https://appleid.apple.com/auth/revoke` | Must be https. |
| `NUVRO_TOKEN_SEALING_KEY` | empty | Base64 of exactly 32 bytes. Seals stored refresh tokens. |
| `NUVRO_CURSOR_SECRET` | empty | Base64, at least 32 bytes. Signs pagination cursors. |
| `NUVRO_SESSION_SIGNING_KEY` | empty | Base64, at least 32 bytes. Signs session access tokens. Required outside development; without it in development, `/v1/auth/sessions` answers 503 and only raw identity tokens authenticate. |
| `NUVRO_DEV_AUTH_CREDENTIAL` | empty | Refused unless `NUVRO_ENV` is `development`. |
| `NUVRO_DEV_AUTH_SUBJECT` | `local-traveller` | |

The four revocation settings are required together outside development. Apple's account
deletion guidance asks apps to revoke Sign in with Apple tokens when an account is deleted, and
a deployment that cannot do that would quietly fail that obligation. Startup refuses instead.

## Client address

See [security standards](../standards/security.md) for how these are used.

| Variable | Default | Notes |
| --- | --- | --- |
| `NUVRO_TRUSTED_PROXY_HOPS` | `0` | Trusted proxies in front of NUVRO that append to `X-Forwarded-For`. |
| `NUVRO_TRUSTED_PROXY_CIDRS` | empty | Comma separated CIDRs of those proxies. Required when hops is above zero. |

## Admission

| Variable | Default | Notes |
| --- | --- | --- |
| `NUVRO_RATE_LIMIT_ENABLED` | `true` | Cannot be `false` outside development. |
| `NUVRO_RATE_LIMIT_STANDARD_READ_BURST` | `120` | Per authenticated traveller. |
| `NUVRO_RATE_LIMIT_STANDARD_READ_PER_SECOND` | `4` | |
| `NUVRO_RATE_LIMIT_EXPENSIVE_SEARCH_BURST` | `12` | Search, generation, replanning, translation. |
| `NUVRO_RATE_LIMIT_EXPENSIVE_SEARCH_PER_SECOND` | `0.2` | |
| `NUVRO_RATE_LIMIT_AUTHENTICATION_BURST` | `300` | Per client address, before a principal exists. |
| `NUVRO_RATE_LIMIT_AUTHENTICATION_PER_SECOND` | `10` | |
| `NUVRO_RATE_LIMIT_COMMERCE_WRITE_BURST` | `20` | |
| `NUVRO_RATE_LIMIT_COMMERCE_WRITE_PER_SECOND` | `0.5` | |

## Providers

| Variable | Default | Notes |
| --- | --- | --- |
| `GOOGLE_MAPS_API_KEY` | empty | Absent means the Google adapters are not constructed at all. |
| `NUVRO_PLACES_BASE_URL` | `https://places.googleapis.com` | |
| `NUVRO_ROUTES_BASE_URL` | `https://routes.googleapis.com` | |
| `NUVRO_WEATHER_BASE_URL` | `https://weather.googleapis.com` | |
| `NUVRO_PROVIDER_TIMEOUT` | `5s` | Between 1s and 30s. |
| `NUVRO_REFERENCE_CURRENCY` | `USD` | ISO 4217. |

An unconfigured provider is reported as unavailable. It is never simulated.

## Background work

| Variable | Default | Notes |
| --- | --- | --- |
| `NUVRO_ALERT_INTERVAL` | `5m` | Between 1m and 1h. |
| `NUVRO_PRICE_WATCH_INTERVAL` | `15m` | Between 1m and 6h. |
| `NUVRO_RETENTION_INTERVAL` | `6h` | Between 1m and 24h. |
