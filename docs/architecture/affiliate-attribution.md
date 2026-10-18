# Affiliate attribution

```mermaid
flowchart TD
    offer[Ranked offer]
    handoff[Server-created handoff]
    allowlist[Partner domain allowlist]
    click[Click recorded once]
    partner[Partner checkout]
    webhook[Signed partner webhook]
    claim[Event claimed once]
    conversion[Normalised conversion]

    offer --> handoff --> allowlist --> click --> partner --> webhook --> claim --> conversion
```

## The client never builds a link

Affiliate URLs are constructed server-side, from the offer the server itself cached. A client
supplies only an opaque offer identifier. There is no endpoint that accepts a redirect target,
so there is no open redirect to exploit.

Every destination is checked against the partner's allowlisted domains before a handoff is
created and again before it is followed. The check requires https, rejects userinfo in the
URL, and matches the host exactly or as a subdomain, so `partner.invalid.evil.invalid` fails.
A fuzz test drives arbitrary strings through it and asserts that anything accepted parses as
https, carries no credentials, and resolves to an allowlisted host.

Handoffs expire after thirty minutes. Following an expired handoff, or one belonging to
another traveller, returns not found.

## Attribution identifiers

Each handoff carries a freshly generated opaque token. No email address, name, passport
country or raw user identifier is ever placed in a partner URL. The token is stored server
side and is never returned to the client.

Only the minimum needed for the partner operation is shared. Passport data, accessibility
preferences and the traveller's wider preference profile are never sent to a partner.

## Clicks

A click is recorded once per handoff, enforced by a unique constraint on the handoff. Repeated
opens do not inflate the count. NUVRO records the handoff, partner, product and time, and
nothing about browsing behaviour beyond that. No third-party mobile tracking SDK is used;
first-party server-generated handoff identifiers do the whole job.

## Webhooks

Verification belongs to the adapter, because every partner signs differently. The shared HMAC
verifier compares a hex digest over the exact payload in constant time. An unsigned request,
a malformed signature, a signature over different content, and an unconfigured secret are all
refused. There is no path that accepts an unsigned production callback.

Payload reads are bounded. Logs record the partner, event type, normalised status and result.
Secrets, raw payloads and booking references are never logged, and references are masked to
their last four characters anywhere they surface.

## Idempotency

A partner may deliver the same event many times. Each delivery is claimed by inserting
`(partner, provider_event_id)` under a unique constraint; a losing insert means the event was
already handled and returns the stored conversion rather than creating a second one.

Conversions themselves upsert on `(partner, provider_conversion_id)`, so one booking is always
one row. Statuses are ranked, so a late-arriving `booking_created` cannot reopen a booking
that has already been cancelled or refunded.

## Accounting

Commission and gross booking value are stored in the partner's original currency with integer
minor units. Nothing is converted into a reference currency at rest; that is a reporting
decision for later, and doing it at write time would destroy the original figure. Commission
never appears in any client response.
