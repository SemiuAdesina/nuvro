# iOS UI completeness matrix

The canonical audit of interaction-state coverage for every screen. `YES` means the state is
implemented with real UI and, where behaviour-bearing, tests. `N/A` means the state cannot occur
on that screen. `PENDING` always carries a reason. Columns: default, loading, refreshing, empty,
partial, error, offline, stale, permission denied, sheet, dialog, destructive confirmation,
success feedback, accessibility, localization, deep-link entry.

Shared behaviour that applies to every screen and is therefore not repeated per row:

- Offline: the root `NuvroOfflineBanner` appears over every main screen (online, offline,
  reconnecting). Screens listed `YES` under offline also degrade their own content.
- Accessibility: `typeStyle` scales with Dynamic Type through `@ScaledMetric` (capped at 1.6×),
  every decorative gradient is a placeholder with alt text through `DynamicDestinationImage`,
  match badges read as one element ("94 percent match"), motion respects Reduce Motion through
  `Motion` and the transient banner, and every custom control carries a label or trait.
- Localization: every string lives in a `NuvroLocalization` copy struct resolved by `Language`.
  Only English values exist today (see `localization` PENDING reasons).
- Appearance: onboarding and authentication are dark by design. Product surfaces follow the
  Appearance setting in 31 Settings (System, Light, Dark) through `ThemePreference`. Every
  colour is a `Palette` token; semantic tokens (`background`, `surface`, `surfaceElevated`,
  `label`, `labelAlpha`, `separator`, `fill`) and the accents resolve per scheme through
  `Color.adaptive`, brand fills (`navy`, `cloud`, `twilightDeep`) stay fixed for hero cards and
  photo scrims. No view carries a hex literal; the scannable ticket keeps a true white field.
- State restoration: each tab owns its own `NavigationStack` path, so switching tabs keeps the
  other tabs' stacks. Onboarding completion and trip start persist in app storage; the session
  persists in the Keychain; picker selections live in `@Observable` models that outlive sheets.
- Previews: `StatePreviews`, `AuthenticationPreviews`, `DiscoverPreviews`, `PlanPreviews` and
  `TravelPreviews` render loading, empty, error, offline, denied and accessibility-size variants
  from in-module fixtures used only by `#Preview`.
- Deep links: navigation is typed (`AppRoute`, `TripRoute`, `SavedRoute`, `ProfileRoute`,
  `AuthenticationRoute`) so any screen can be entered by pushing its route; no URL scheme is
  registered yet.

Legend for PENDING reasons:

- **P1** Live data source: the screen renders fixture content synchronously; the backend endpoint
  does not exist in the OpenAPI contract, so loading, refreshing, error and stale are wired only
  where a `LoadState` model exists and previews or tests exercise them.
- **P2** Translations: copy structs exist with English only; other languages resolve to English
  until the localization workflow lands.
- **P3** URL scheme and universal links are not registered; typed routes exist.

| Screen | default | loading | refreshing | empty | partial | error | offline | stale | permission denied | sheet | dialog | destructive | success | accessibility | localization | deep-link |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 00 Splash | YES | YES (bootstrap restoring state) | N/A | N/A | N/A | N/A | YES (restore keeps stored session) | N/A | N/A | N/A | N/A | N/A | N/A | YES | PENDING P2 | N/A |
| 01 Onboarding Discover | YES | N/A | N/A | N/A | N/A | N/A | YES | N/A | N/A | YES (language) | N/A | N/A | N/A | YES | YES (6 languages) | N/A |
| 02 Onboarding Plan | YES | N/A | N/A | N/A | N/A | N/A | YES | N/A | N/A | N/A | N/A | N/A | N/A | YES | YES | N/A |
| 03 Onboarding Travel | YES | N/A | N/A | N/A | N/A | N/A | YES | N/A | N/A | N/A | N/A | N/A | N/A | YES | YES | N/A |
| AUTH-01 Landing | YES | YES (authenticating) | N/A | N/A | N/A | YES | YES | N/A | N/A | N/A | N/A | N/A | YES (session established) | YES | PENDING P2 | YES (route) |
| AUTH-02 Email | YES | YES (submitting) | N/A | N/A | N/A | YES (format, backend, rate limit) | YES | N/A | N/A | N/A | N/A | N/A | YES (code sent) | YES | PENDING P2 | YES (route) |
| AUTH-03 Verification | YES | YES (verifying, resending) | N/A | N/A | N/A | YES (wrong, expired, rate limit) | YES | N/A | N/A | N/A | N/A | N/A | YES | YES (single semantic field, AutoFill) | PENDING P2 | YES (route) |
| AUTH-04 Failure | YES | N/A | N/A | N/A | N/A | YES | YES | N/A | N/A | N/A | N/A | N/A | N/A | YES | PENDING P2 | N/A |
| SETUP-01 Welcome | YES | N/A | N/A | N/A | N/A | N/A | YES | N/A | N/A | N/A | N/A | N/A | N/A | YES | PENDING P2 | N/A |
| SETUP-02 Residence | YES | N/A | N/A | YES (nothing chosen) | N/A | YES (continue blocked) | YES | N/A | N/A | YES (searchable country) | N/A | N/A | N/A | YES (flag and name announced together) | PENDING P2 (names from Foundation) | N/A |
| SETUP-03 Passports | YES | N/A | N/A | YES | N/A | YES (blocked unless declined) | YES | N/A | N/A | YES (searchable country) | N/A | YES (remove passport, reversible) | N/A | YES | PENDING P2 | N/A |
| SETUP-04 Permits | YES | N/A | N/A | YES | N/A | N/A | YES | N/A | N/A | YES (country or Schengen) | N/A | N/A | N/A | YES | PENDING P2 | N/A |
| SETUP-05 Airport | YES | N/A | N/A | YES | N/A | YES (blocked) | YES | N/A | N/A (location never required) | YES (searchable airport) | N/A | N/A | N/A | YES | PENDING P2 | N/A |
| SETUP-06 Currency | YES | N/A | N/A | YES (suggested, never forced) | N/A | YES (blocked) | YES | N/A | N/A | YES (searchable ISO currencies) | N/A | N/A | N/A | YES (name and code announced together) | PENDING P2 | N/A |
| SETUP-07 Paperwork | YES | N/A | N/A | N/A | N/A | YES (blocked) | YES | N/A | N/A | N/A | N/A | N/A | N/A | YES | PENDING P2 | N/A |
| SETUP-08 to 15 Preferences | YES | N/A | N/A | N/A | N/A | YES (required steps blocked) | YES | N/A | N/A | N/A | N/A | N/A | N/A | YES (Reduce Motion fades instead of moving) | PENDING P2 | N/A |
| SETUP-16 Review | YES | N/A | N/A | N/A | N/A | N/A | YES | N/A | N/A | N/A | N/A | N/A | YES (profile saved, draft cleared) | YES | PENDING P2 | N/A |
| 05 Discover Home | YES | YES | YES (pull) | YES | YES (cached with refresh indicator) | YES (retry) | YES | N/A | N/A | N/A | N/A | N/A | YES (Saved) | YES | PENDING P2 | YES |
| 06 Match Setup | YES | N/A | N/A | N/A | N/A | YES (missing departure blocks submit) | YES | N/A | N/A | YES (airport, dates, travellers) | N/A | N/A | N/A | YES | PENDING P2 | YES |
| 07 Match Results | YES | YES | N/A | YES | N/A | YES (retry) | YES | N/A | N/A | N/A | N/A | N/A | YES (saved, compare) | YES | PENDING P2 | YES |
| 08 Destination Detail | YES | PENDING P1 | PENDING P1 | N/A | N/A | PENDING P1 | YES | N/A | N/A | N/A | N/A | N/A | YES (saved) | YES | PENDING P2 | YES |
| 09 Trip Plan Hub | YES | N/A | N/A | N/A | N/A | N/A | YES | N/A | N/A | YES (dates, travellers, budget) | YES (edit trip) | N/A | YES (trip updated) | YES | PENDING P2 | YES |
| 10 Full Itinerary | YES | N/A | N/A | YES (empty day copy) | N/A | N/A | YES | N/A | N/A | YES (move) | YES (context menu) | YES (remove reservation) | YES (removed, moved) | YES | PENDING P2 | YES |
| 11 Budget | YES | N/A | N/A | N/A | N/A | N/A | YES | N/A | N/A | N/A | N/A | N/A | N/A | YES | PENDING P2 | YES |
| 12 Stay | YES | PENDING P1 | PENDING P1 | PENDING P1 | N/A | PENDING P1 | YES | N/A | N/A | N/A | N/A | N/A | YES (saved) | YES | PENDING P2 | YES |
| 13 Travel Mode | YES | N/A | N/A | N/A | N/A | N/A | YES (weather shows saved copy) | YES (weather freshness) | N/A | N/A | N/A | N/A | YES (plan updated) | YES | PENDING P2 | YES |
| 14 Live Replanning | YES | YES (applying) | N/A | N/A | N/A | N/A | YES | N/A | N/A | N/A | N/A | N/A | YES | YES | PENDING P2 | YES |
| 15 Language Companion | YES | YES (listening) | N/A | N/A | N/A | N/A | YES (offline phrases) | N/A | YES (microphone, camera) | YES (permission) | N/A | N/A | N/A | YES | PENDING P2 | YES |
| 16 Explore Nearby | YES | N/A | N/A | YES (no results) | N/A | N/A | YES | N/A | YES (location, destination centre fallback) | YES (permission) | N/A | N/A | YES (added to today) | YES | PENDING P2 | YES |
| 17 Tickets & Documents | YES | N/A | N/A | YES | YES (expired, no barcode) | YES (barcode unavailable) | YES (offline copies) | N/A | N/A | N/A | YES (context menu) | YES (remove download) | YES (downloaded, removed) | YES | PENDING P2 | YES |
| 18 Smart Alerts | YES | YES (refreshing) | YES (pull) | YES | YES (unread, read, dismissed) | N/A | YES | N/A | YES (notifications) | YES (preferences) | N/A | N/A | N/A | YES | PENDING P2 | YES |
| 19 Trip Summary | YES | N/A | N/A | N/A | N/A | N/A | YES | N/A | N/A | N/A | N/A | N/A | N/A | YES | PENDING P2 | YES |
| 20 Trips Library | YES | N/A | N/A | YES | N/A | N/A | YES | N/A | N/A | N/A | YES (overflow) | YES (delete) | YES (archived, deleted, downloaded) | YES | PENDING P2 | YES |
| 21 Saved | YES | N/A | N/A | YES (destinations, places) | N/A | N/A | YES | N/A | N/A | N/A | N/A | N/A | N/A | YES | PENDING P2 | YES |
| 22 Compare | YES | PENDING P1 | N/A | N/A | YES (metric unavailable keeps the rest) | PENDING P1 | YES | N/A | N/A | N/A | N/A | N/A | YES (chosen) | YES | PENDING P2 | YES |
| 23 Flights | YES | YES | YES (pull) | YES | YES (cached fares) | YES (provider unavailable, offline) | YES | YES (stale fares) | N/A | YES (partner hand-off) | N/A | N/A | N/A | YES | PENDING P2 | YES |
| 24 Things to Do | YES | PENDING P1 | PENDING P1 | YES | N/A | PENDING P1 | YES | N/A | N/A | N/A | N/A | N/A | YES (added) | YES | PENDING P2 | YES |
| 25 Transport | YES | N/A | N/A | N/A | N/A | N/A | YES | N/A | N/A | N/A | N/A | N/A | N/A | YES | PENDING P2 | YES |
| 26 Packing | YES | N/A | N/A | N/A | N/A | N/A | YES | N/A | N/A | N/A | YES (remove item) | N/A | YES (section done, all packed) | YES | PENDING P2 | YES |
| 27 Group Planning | YES | N/A | N/A | N/A | N/A | N/A | YES | N/A | N/A | YES (invite) | YES (member actions) | YES (remove, leave) | YES (removed, invited) | YES | PENDING P2 | YES |
| 28 Practical Info | YES | N/A | N/A | N/A | N/A | N/A | YES | YES (entry checked label) | N/A | N/A | N/A | N/A | N/A | YES | PENDING P2 | YES |
| 29 Profile | YES | N/A | N/A | N/A | N/A | N/A | YES | N/A | N/A | YES (preference pickers) | N/A | N/A | YES (preference updated) | YES | PENDING P2 | YES |
| 30 Language & region | YES | N/A | N/A | N/A | N/A | N/A | YES | N/A | N/A | YES (five pickers) | N/A | N/A | N/A | YES | PENDING P2 | YES |
| 31 Settings | YES | YES (deleting) | N/A | N/A | N/A | YES (delete failed inline) | YES | N/A | YES (live permission states, Open Settings) | YES (delete explainer) | YES (sign out) | YES (sign out, delete) | N/A | YES | PENDING P2 | YES |

## Remaining PENDING summary

| Reason | Screens | What unblocks it |
| --- | --- | --- |
| P1 live data source | 08, 12, 22 (loading and error), 24 | Discover, destination, stay, comparison and experience endpoints in the OpenAPI contract. The `LoadState` pattern from 05 and 23 is the template. |
| P2 translations | all product screens | Translated copy values in `NuvroLocalization`; the structure resolves by `Language` today. |
| P3 URL scheme | all | Registering a scheme and mapping URLs to the existing typed routes. |
