# Offline trip pack

A bounded snapshot so a traveller with no signal still knows where they are going.

## Contents

- Trip identity, destination timezone and the itinerary version it was built from
- The full day-by-day itinerary, with times and places
- Essential phrases, including emergency phrases
- Currently visible alerts
- A generation time, an expiry and a freshness record

## Nothing here is live

The pack's freshness status is always `cached`, never `live`. Weather, route times and opening
hours drift the moment the pack is generated, so the pack carries a note saying its weather was
cached at generation and does not restate route timings as current.

## Deliberate exclusions

The pack declares what it left out, so the client can explain the gap rather than appear
broken:

| Excluded | Reason |
| --- | --- |
| `provider_photography` | Provider licensing does not permit redistribution or offline retention |
| `provider_map_tiles` | Same, and a tile pack is a separate product decision |
| `live_route_times` | Traffic-dependent values are meaningless once stored |

## Versioning

Each pack carries its own version, the itinerary version it was built from, a generation time
and a 24-hour expiry. The client compares the itinerary version against the live trip to decide
whether to refresh, rather than downloading blindly.

Generation is synchronous because it reads only stored state: the itinerary, the phrase
catalog and visible alerts. No provider is called, so there is no reason for a job or a polling
protocol.
