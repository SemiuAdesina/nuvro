# Provider gap report

Phase 2 implements the Discover intelligence flow against consumer-owned interfaces. No
external provider is integrated, no credential is configured, and no provider is faked. With
nothing configured the engines return honest `unknown` and `insufficient_evidence` evidence,
and the API still answers with structured recommendations.

Test fakes exist only in `_test.go` files. Nothing in runtime code invents a visa rule, a
flight price, an exchange rate or an image.

## Gaps

| Capability | Interface | Why it is needed | Candidates | Credentials | Production ready | Phase |
| --- | --- | --- | --- | --- | --- | --- |
| Entry requirements | `engine/feasibility.EntryRequirementProvider` | Access assessment is legally consequential and must come from an authoritative, current source | IATA Timatic, Sherpa | None | No | 4 |
| Flight evidence | `engine/feasibility.FlightEvidenceProvider` | Reachability, stops, duration and indicative price feed both feasibility and match | Amadeus Self-Service, Duffel, Kiwi Tequila | None | No | 4 |
| FX rates | `engine/feasibility.FXRateProvider`, `engine/budget.Rates` | Cost estimates must be normalised into the traveller's currency with provenance | European Central Bank reference rates, Open Exchange Rates, Fixer | None | No | 4 |
| Trip cost evidence | `engine/feasibility.TripCostEstimator`, `engine/budget.EstimateProvider` | Affordability and the trip budget both need a defensible cost range, not a guess | Numbeo, internal model over stay and activity providers | None | No | 4 |
| Intent interpretation | `engine/understanding.Interpreter` | Optional free-text discovery input | OpenAI, Anthropic | `OPENAI_API_KEY` exists in the sample environment but is unset and unused | No | 4 |
| Explanation | `engine/recommendation.ExplanationGenerator` | Human-readable summary over already-computed facts | OpenAI, Anthropic | As above | No | 4 |
| Destination media | `engine/media.CandidateReader` | S01 supplies photography for destination cards | Google Place Photos | Adapter written; needs `GOOGLE_MAPS_API_KEY` | Adapter ready, unverified against the live service | 4 |
| Places | `engine/places.SearchProvider`, `engine/places.DetailsProvider` | Real establishments, coordinates, categories and opening evidence for E04, E05 and nearby search | Google Places API (New) | Adapter written; needs `GOOGLE_MAPS_API_KEY` | Adapter ready, unverified against the live service | 4 |
| Routes | `engine/routing.Provider` | Travel time and distance for itinerary feasibility | Google Routes API | Adapter written; needs `GOOGLE_MAPS_API_KEY` | Adapter ready, unverified against the live service | 4 |
| Weather | `engine/weather.Provider` | Current conditions and forecast inside the provider horizon | Google Weather API | Adapter written; needs `GOOGLE_MAPS_API_KEY` | Adapter ready, unverified against the live service | 4 |
| Translation | `engine/language.Provider` | E10 cannot translate anything without it; the endpoint returns unavailable today | Google Cloud Translation, DeepL, an approved model provider | None | No | 5 |
| Push delivery | `adapter/external/push.Sender` and `push.Registry` | Alerts are stored and served, but nothing reaches a locked phone | APNs for iOS, FCM for Android later | No certificate or key configured | No | 5 |
| Device tokens | none yet | Deliberately not built: token storage belongs with the push adapter, including rotation and revocation | Not applicable | Not applicable | No | 5 |
| Destination catalog | `usecase/discover.DestinationCatalog` | The development seed is fourteen destinations, not a global catalog | Licensed geographic dataset, or an editorial catalog | Not applicable | No | 4 |

## Commerce partners

E13 is implemented against consumer-owned interfaces. No commerce partner is approved,
credentialed or integrated, so every category below returns an empty result with a partner
status rather than a fake offer. Nothing is scraped and no provider was signed up for.

| Category | Interface | Candidate partners | Affiliate approval | Credentials | Webhooks | Production ready | Phase |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Flights | `engine/commerce.FlightOfferProvider` | Duffel, Kiwi Tequila, Skyscanner affiliate | Not applied for | None | Unknown until approval | No | 6 |
| Stays | `engine/commerce.StayOfferProvider` | Booking.com affiliate, Expedia Rapid, Hotelbeds | Not applied for | None | Unknown until approval | No | 6 |
| Activities | `engine/commerce.ActivityOfferProvider` | GetYourGuide, Viator, Tiqets | Not applied for | None | Unknown until approval | No | 6 |
| Transfers | `engine/commerce.TransferOfferProvider` | Welcome Pickups, Kiwitaxi | Not applied for | None | Unknown until approval | No | 6 |
| Cars | Interface pending its first slice | Rentalcars, Discover Cars | Not applied for | None | Unknown until approval | No | 6 |
| eSIM | Interface pending its first slice | Airalo, Holafly | Not applied for | None | Unknown until approval | No | 6 |
| Insurance | Interface pending its first slice | Heymondo, SafetyWing | Not applied for; high-consequence, needs legal review | None | Unknown until approval | No | 6 |
| Premium entitlement | `usecase/commerce.Entitlements` | StoreKit server notifications and receipt validation | Not built | None | Not applicable | No | 6 |

Insurance carries an extra requirement: NUVRO may surface provider-supplied product facts and
hand off, but must never claim a policy is legally sufficient, guarantee coverage, invent
exclusions, or give personalised insurance advice.

Car, eSIM and insurance interfaces are deliberately not written yet. Adding an interface with
no consumer and no slice would be speculative scaffolding; each will be defined by the first
real use of it, exactly as the other categories were.

## Selection requirements

Before any of these is integrated, record the licence, commercial usage terms, rate limits,
availability guarantee, data freshness, coverage and personal-data handling. Entry
requirements additionally need an accuracy and liability review, since a wrong answer can
strand a traveller. None of these may be chosen casually or introduced silently.

## Google adapters

Phase 3 added real adapters for Google Places, Routes and Weather, plus Place Photos for S01.
Their response parsing, unit conversion, field masks, attribution handling and error mapping
are covered by tests against sanitized fixtures and local test servers. No automated test
calls Google. They have not been exercised against the live service, because no credential is
configured here, so they are not production ready and are reported as such.

Set `GOOGLE_MAPS_API_KEY` to enable them. Before relying on them in production, review Places
and Photos licensing and caching terms, the attribution display requirements, per-endpoint
quotas, and the cost controls in [provider-cost-controls.md](provider-cost-controls.md).

## Behaviour while unconfigured

- Access is `unknown` and says so, with the reason recorded.
- Affordability is `insufficient_evidence` and returns no estimate.
- Flight is `unknown`, distinct from a route that genuinely does not exist.
- Overall feasibility is `unknown`; nothing is excluded on missing evidence alone.
- The match score still computes, with partial credit for unknowns.
- Ranking still orders deterministically.
- Destination sections report `unavailable` with a reason instead of empty content.
- Weather beyond the provider horizon reports `beyond_forecast_horizon`, never a guess.
- Route evidence is unavailable with no duration rather than an estimate.
- Itinerary generation fails with a clear message rather than inventing places.
- The trip budget returns no estimate and unknown confidence.
- Translation returns an unavailable error rather than an invented translation.
- Phrases are returned untranslated and explicitly marked as such.
- Alerts are still evaluated, stored and served; only push delivery is missing.
- Travel Mode reports `route_unavailable` and omits any leave time.
- Commerce searches return an empty offer list with per-partner status, never a fabricated offer.
- Handoffs cannot be created because no offer can be cached without a partner.
- Price monitoring reports unavailable, because no entitlement backend and no repricer exist.
