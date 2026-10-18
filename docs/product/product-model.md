# Product model

NUVRO is a global vacation intelligence and travel companion platform.

**The world, made personal.**

## The flow

```
DISCOVER  →  PLAN  →  TRAVEL
```

A traveller who does not know where to go starts at Discover. A traveller who already has a
destination starts at Plan. A traveller who has arrived is in Travel. The product does not
end when the trip is booked.

## The global foundation

Every part of the product sits on this foundation. None of it is a settings screen
afterthought; each of these changes what the product returns.

```
Language              Locale                Currency
Home country          Departure airport     Units
Timezone              Passport and nationality context
Accessibility         Profile               Cloud sync
```

Language and locale are separate concepts. Passport and nationality context is separate from
current location. A traveller can be a Nigerian passport holder, living in the United Kingdom,
currently in Japan, reading in French, paying in euros. All five facts matter and none of them
substitutes for another.

## The four signature capabilities

| Capability | What it answers |
| --- | --- |
| **Vacation Match Score** | How well does this destination fit this traveller, and why? |
| **Where Can My Money Take Me?** | Given this budget, these dates and this departure airport, where can I actually go? |
| **Everything About This Place** | What do I need to know about this destination, gathered and current? |
| **Adaptive Travel Mode** | My plan just broke. What now? |

## Feature systems

Each has a home in the architecture. None is implemented.

| System | Backend home | Client home |
| --- | --- | --- |
| Global foundation | `domain/locale`, `domain/user` | `NuvroLocalization`, `NuvroProfile` |
| Discover | `usecase/discover`, `engine/understanding`, `engine/recommendation` | `NuvroDiscover` |
| Destination intelligence | `usecase/destination`, `engine/intel` | `NuvroDestination` |
| Plan | `usecase/plan`, `engine/itinerary`, `engine/budget` | `NuvroPlan` |
| Travel mode | `usecase/travel`, `engine/replanning`, `engine/routing` | `NuvroTravel` |
| Language companion | `engine/language` | `NuvroTravel` |
| Smart alerts | `usecase/alerts`, `worker/jobs` | `NuvroTravel`, `NuvroPlan` |
| Booking and affiliate | `usecase/commerce`, `engine/commerce` | across features, always labelled |
| Trust, privacy and quality | cross-cutting | cross-cutting |
| Premium | `usecase/profile` | `NuvroProfile` |

## Commerce independence

Commerce sits underneath the experience. It does not steer it.

**Affiliate payout never influences Vacation Match Score or organic recommendation ranking.**

Commerce attaches booking options and attribution to results that have already been ranked.
Sponsored placement, if ever introduced, is explicitly labelled and separated from organic
results. A traveller must be able to trust that a high Match Score means fit, not margin.

## What the backend owns

Canonical travel intelligence lives in Go so that iOS and Android return identical results
for the same request:

```
Vacation Match Score          destination ranking
travel intelligence           budget calculation rules
itinerary orchestration       provider aggregation
canonical recommendation      entry requirement reasoning over trusted facts
affiliate independence rules
```

## What the clients own

```
presentation        platform integration      local UI state
offline cache       navigation                accessibility
local formatting    device capabilities
```

A client formats 125.50 US dollars from `amount_minor` and `currency`. It does not decide
whether Lisbon outranks Porto.
