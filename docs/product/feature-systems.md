# Feature systems

None of this is implemented. This document records where each system will live so that the
first commit in each area lands in the right place.

## Global foundation

Language, locale, currency, home country, departure airport, units, timezone, passport and
nationality context, accessibility and cloud sync.

The backend stores canonical values and canonical units. Clients convert for display:
kilometres and miles, Celsius and Fahrenheit, 12-hour and 24-hour, currency and date formats.
UI copy is never stored inside a canonical business model.

## Discover

Turns a traveller's intent into ranked destinations. Backed by Vacation Understanding,
Recommendation and Vacation Match. Includes Where Can My Money Take Me?, which is a budget
constrained Discover query rather than a separate product.

## Destination intelligence

Everything About This Place. Assembles current facts about a destination from trusted
providers, with source and observation time preserved, and lets AI summarise and explain what
was gathered. The AI never adds a fact that no provider supplied.

## Plan

Itinerary structure, sequencing and feasibility, with trip cost and budget. A plan knows what
is fixed, what is flexible and what is at risk.

## Travel mode

Adaptive Travel Mode. When a flight moves, a venue closes or weather turns, the plan is
repaired rather than abandoned. Route and time, live replanning and weather intelligence work
together here.

## Language companion

Translation and phrase context in the traveller's language, aware of where they are and what
they are doing.

## Smart alerts

Scheduled and event-driven checks that run in the worker rather than on a request: price
movement, requirement changes, weather deterioration, transport disruption. Delivery respects
the traveller's notification preferences and quiet hours.

## Booking and affiliate

Hotels, flights, activities, cars, airport transfers, eSIM, travel insurance, cruises and
destination services. The architecture supports affiliate attribution, click tracking,
conversion tracking, commission reconciliation and partner disclosure.

Independence from organic ranking is the invariant that makes this system acceptable at all.

## Trust, privacy and quality

Freshness on every dynamic value, source attribution on high-consequence facts, honest
handling of gaps, and no dark patterns around commerce. Entry requirements always carry
source, retrieval time, traveller nationality context, destination and travel dates.

## Premium

Subscription entitlement gating. Entitlement is checked in a use case, never in a view.
