# 0008. First-party sessions instead of identity tokens as bearer credentials

Date: 2026-09-09

Status: Accepted

## Context

Every product operation authenticates with a bearer credential. Until now that credential was
the Sign in with Apple identity token itself, verified against Apple's keys on every request.
Apple issues that token with a ten minute expiry, and there is no way for an app to obtain a
fresh one without putting the sign-in sheet in front of the traveller again. A client that
used the identity token directly would stop working ten minutes after sign-in.

The iOS client already modelled a session as an access token, a refresh token and an expiry.
The backend had no operation that produced one.

## Decision

The backend issues its own sessions. `POST /v1/auth/sessions` verifies an identity credential
once, resolves or enrols the account, links the Sign in with Apple authorization code when
present, and returns a short-lived HS256 access token plus a single-use refresh token.
`POST /v1/auth/sessions/refresh` rotates the pair; presenting a refresh token that has already
been used revokes the whole session family, on the assumption that one of the two holders is
not the traveller. `DELETE /v1/auth/sessions` revokes the session named by the access token.

The access token lives fifteen minutes and is checked against the session row on every
request, so sign-out and account closure take effect immediately rather than at expiry. The
refresh token lives ninety days and is stored only as a SHA-256 hash.

Identity tokens remain accepted as bearer credentials. The authenticator tries the session
verifier first and falls back to the identity verifier, so the development credential and a
freshly minted Apple token keep working for tooling and for the contract suite.

## Consequences

Mobile clients exchange the Apple credential once and never send it again. Sessions need
`NUVRO_SESSION_SIGNING_KEY`, required outside development. A sessions table joins the schema
and is purged with the account. Refresh-token reuse is a security signal, not a retry, and the
client must treat a refused refresh as a sign-out.
