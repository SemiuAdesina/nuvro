# NUVRO iOS

Native iPhone application. Swift, SwiftUI and Apple platform APIs. No React Native, no
Flutter, and no web view as the primary interface architecture.

## Current state

`Packages/NuvroKit` is a real Swift package that builds and tests today.

```bash
cd clients/ios/Packages/NuvroKit && swift test
```

Four modules exist:

| Module | What is in it |
| --- | --- |
| `NuvroModels` | `Money` and `APIFailure`, decoded from the exact JSON shapes in `contracts/openapi`. |
| `NuvroDesignSystem` | Palette, sky gradients, typography, radii, motion curves, Liquid Glass surfaces, the wordmark with its route arc, and `DesignCanvas`. |
| `NuvroLocalization` | The ten picker languages and onboarding copy in English, Spanish, French, Portuguese, German and Japanese. |
| `NuvroOnboarding` | Splash, three screens, language sheet, the 2.5D world, camera poses and the flow model. |
| `NuvroDiscover` | Discover Home, Vacation Match Setup and Vacation Match Results, with their models. |
| `NuvroDestination` | Destination Detail: hero, glance facts, why it fits, guide categories, areas, activities, MapKit preview. |
| `NuvroPlan` | Trip Plan Hub, Full Itinerary, Budget and Trip Cost, and Stay with list and map. |
| `NuvroTravel` | The active trip: Travel Mode, Live Replanning, Language Companion, Explore Nearby, Tickets and Documents, Smart Alerts, Trip Summary. |
| `NuvroSaved` | The Saved tab with destination selection and the two-column Compare screen. |
| `NuvroProfile` | Profile and preferences, Language & region with unit preview, and Settings with sign out and account deletion. |
| `NuvroAuthentication` | Sign in with Apple, passwordless email code, Keychain session store, the coordinator state machine and the bootstrap state. See `docs/ios/authentication-flow.md`. |
| `NuvroPermissions` | Permission kinds and states, the system authority (CoreLocation, UserNotifications, AVFoundation), the explainer and denied presentation, and the iOS Settings link. |
| `NuvroConnectivity` | `ConnectivityMonitor` over `NWPathMonitor` with online, offline and reconnecting states. |
| `NuvroTravelSetup` | The one-question-at-a-time Travel Setup (SETUP-01 to 16): residence, passports, permits, departure airport, currency, paperwork preference and travel preferences, with draft persistence, review editing, and single-step edit mode reused by Profile. |

## Launch order and Travel Setup

`LaunchPhase` in `NuvroAuthentication` decides what the root shows: restore the session, then
authentication, then Travel Setup for a signed-in traveller without a saved `TravelProfile`,
then the story screens (started at the first story stage so the splash is not shown twice),
then the main tabs. Residence, passport countries, departure airport and currency are separate
answers and are never inferred from each other. The client collects preferences only; entry
eligibility, affordability and access are backend engines (E11, E12, E15).

The chosen currency is the primary display currency: Match Setup budgets use it with a
currency-appropriate scale from prototype content, and Language & region starts from it. Trip
estimates in fixtures remain in their own currency until the backend returns converted amounts;
the client never invents exchange rates.

## Appearance

Settings offers System, Light and Dark (`ThemePreference`). Colours come only from `Palette`:
semantic tokens resolve per scheme through `Color.adaptive`, accents brighten slightly in dark,
and brand fills stay fixed. Onboarding and authentication are always dark by design.

## Supporting states

`NuvroDesignSystem` carries the shared state primitives: `LoadState` and `LoadFailure` (in
`NuvroModels`), `NuvroLoadingState`, `NuvroEmptyState` (built on `ContentUnavailableView`),
`NuvroErrorState`, `NuvroInlineError`, `FreshnessLabel`, `NuvroOfflineBanner`,
`TransientFeedback` with `TransientBanner`, `NuvroPermissionExplainer`, `NuvroPickerRow`, and the
picker sheets under `Presentation/`: `OptionPickerSheet`, `MultiOptionPickerSheet`,
`SearchablePickerSheet`, `CountPickerSheet` and `DateSelectionSheet`.

Two documents are the canonical audit: `docs/ios/ui-completeness-matrix.md` lists every state
per screen and `docs/ios/presentation-registry.md` lists every sheet, alert, dialog, picker,
permission explainer and banner. A new presentation or state is recorded there in the same change.

Feature views never call the network or a permission API directly. They read `PermissionCenter`,
`ConnectivityMonitor`, `TransientFeedback` and `AuthenticationCoordinator` from the environment,
all created once in `RootView`.

## The app project

The Xcode project is generated, never hand-edited:

```bash
cd clients/ios/App && xcodegen generate
```

`project.yml` is the committed source. `Nuvro.xcodeproj` is a machine-generated file, ignored
by Git, and rebuilt by the `ios` workflow on every run. The app target contains wiring only:
`NuvroApp` and `RootView`, which shows onboarding and, until the Discover module exists,
replays it after the traveller finishes.

## The product screens

Screens 05 to 09 were built from the approved Claude Design canvas with the same method as
onboarding. Every reusable pattern lives in `NuvroDesignSystem`: surface cards, section
headers, tag chips, fact tiles, match badges and the match ring, match reason rows, segmented
choices, count steppers, icon tiles, glass button styles, the call to action, and the action
dock that floats a CTA over scrolling content.

Navigation is native. Tabs are a `TabView` whose iOS 26 tab bar is Liquid Glass by default.
Pushed screens hide it. The results hero zooms into the destination hero with the native
`navigationTransition(.zoom)`, which respects Reduce Motion on its own. Toolbar controls are
plain items so the system's glass styles them; wrapping a custom glass pill inside a toolbar
item produces a doubled pill and clipped text.

Values shown in the product are derived from the models, never typed into views: the Discover
card's "From" price is the match estimate's low bound, readiness counts and budget totals are
computed from the trip plan, and the destination's "why it fits" tiles are derived from the
match, the budget ceiling and the estimate. The Match Score itself is data from the backend
contract and is never computed on the client.

Prototype content for the flow lives in the app target's `PrototypeContent`, typed against the
`NuvroModels` contract shapes. It is replaced by API responses when the Discover endpoints
exist.

## Plan into Travel

Screens 10 to 14 continue the same method. The Trips tab owns the sequence: hub, itinerary,
budget, stay, then Travel Mode once the trip begins, then replanning. Travel Mode swaps the
planning chrome for a dawn wash and a navy "Next" hero; the tab bar returns because Trips is
the active tab.

Timeline rows and travel-time connectors are one component serving both the full itinerary
and Travel Mode's Today list. The budget ring uses the same category colours as the hub's
stacked bar, and both read one set of budget lines, so the totals cannot disagree.

Travel Mode runs on a prototype clock, 09:26 on the third day, so "Leave in 22 min" is computed
from the next activity's time minus the walk from the previous stop, never typed. Applying the
rain adjustment replaces the afternoon activity in both the Travel Mode day and the shared
trip plan, so the full itinerary reflects it afterwards.

## Actions live in the page

Product screens end with their primary action inline, after the last content block, rather
than in a dock pinned over the scroll view. Only onboarding keeps a fixed call to action,
because those screens do not scroll. Tab roots carry no top-right profile button: Profile is a
tab. The List/Map switcher on Stay is the design system's segmented control in its pill form,
with the toolbar's shared glass hidden so the control is not a pill inside a pill.

## Screens 15 to 19 and destination media

Language, Nearby, Wallet, Alerts and Summary all hang off Travel Mode: the quick actions open
Translate, Navigate and Tickets, and the bell in the toolbar opens Trip updates. Alerts route
back into Live Replanning, Travel Mode, or the wallet, and the "Later" alert previews the trip
summary so the sequence stays tappable before the trip ends.

Photography is a backend contract. `DestinationMedia` carries the selected image URL, a
placeholder token, alt text, provider and attribution; `DynamicDestinationImage` renders the
placeholder gradient immediately and crossfades the photo in when it loads. The prototype ships
no image URLs, so every named place shows its placeholder state with attribution text. Provider
selection, quality scoring, seasonal choice and rotation live in the Go engine, never here.

## Screens 20 to 31

The Trips tab now opens on the Trips Library; the hub, itinerary, budget, stay and the planning
modules (Flights, Things to do, Getting around, Get ready, Travelling together, Know before you
go) sit beneath it. Saved owns Compare: selecting two destination cards enables the Compare
pill, and every metric states "▲ better" in words rather than by colour. Profile owns Language &
region, whose unit choices drive a live preview, and Settings. Entry rows on Know before you go
carry the passport caveat and a link to the official source; flights show when the fare was
checked and a prototype-fare disclaimer.

## Screen titles

Titles are native large titles, not hand-drawn text. `screenTitle(_:subtitle:)` in NuvroDesignSystem
applies `navigationTitle`, large display mode and, on iOS 26, `navigationSubtitle`; the system
handles the collapse into the inline bar, Dynamic Type, Reduce Motion and VoiceOver. Trailing
actions (Plan trip, Compare, Alert preferences) live in the toolbar as glass pills. Brand colour
for both title sizes is set once through `NavigationTitleStyling.apply()` at app launch. Photo
heroes (Destination, Trip Summary) and identity headers (Hub, Full Itinerary) keep their own
in-content headers on purpose.

## Onboarding geometry

The design was authored on a 402 by 874 point canvas, which is the iPhone 17 Pro. Every
coordinate in `NuvroOnboarding` is the design's own value, and `DesignCanvas` scales them at
runtime:

- Horizontal values scale by the device width over 402.
- Top-anchored values are measured from the safe area top, with 62 as the design's safe top.
- Bottom-anchored values are measured from the safe area bottom, with 34 as the design's safe
  bottom, so the call to action sits 58 points above the home indicator on every device.

Three parallax layers share one camera. Depth factors are 0.35 for sky, 1.0 for the globe and
1.25 for landmarks, anchored at fifty percent across and forty two percent down. Camera poses
per stage are in `CameraPose.byStage`. Cards, chips and the live map live in an unscaled
overlay so text never blurs.

Reduce Motion pins the camera to the first screen pose and replaces every transform with a
fade. The system setting and an explicit model flag are both honoured.

## Known stand-ins

- The wordmark uses Outfit SemiBold when that font is registered, and falls back to the
  system font with the same sixteen percent tracking. Outfit is not bundled; add it under
  `App/Nuvro/Resources` and register it in the generated Info.plist when the brand is approved.
- Destination photography is the design's labelled gradient stand-in, drawn by
  `PhotoStandIn`. Real photography replaces it screen by screen as the Discover work lands.
- Onboarding copy lives in a typed table rather than a string catalogue, because the language
  picker switches language at runtime before the app locale can change. Migrating to string
  catalogues loaded per language bundle is the planned path once the app locale flow exists.

## Module plan

A target is added to `Package.swift` when its first real source file is written. A declared
target with an empty directory makes the whole package fail to build, so the plan lives here.

| Module | Responsibility |
| --- | --- |
| `NuvroModels` | Contract types. No behaviour beyond decoding, validation and formatting. |
| `NuvroDomain` | Client-side repository protocols and client use cases. Client concerns only. |
| `NuvroNetworking` | HTTP transport, request building, response decoding, error mapping. |
| `NuvroData` | Repository implementations over networking and local storage. |
| `NuvroDesignSystem` | Tokens, typography, components, glass, motion, layout, accessibility. |
| `NuvroLocalization` | String catalogues and language and locale handling. |
| `NuvroLocation` | Location authorisation and updates. |
| `NuvroOnboarding` | Splash, three onboarding screens, language selector, Discover handoff. |
| `NuvroDiscover` | Discover feature. |
| `NuvroDestination` | Destination detail feature. |
| `NuvroPlan` | Planning feature. |
| `NuvroTravel` | Travel mode feature. |
| `NuvroProfile` | Profile and preferences feature. |

Each module gets a matching `Tests/<Module>Tests` target in the same change.

## Dependency direction

```
NuvroModels
    ↑
NuvroDomain  ←  NuvroData  ←  NuvroNetworking

NuvroDesignSystem   NuvroLocalization   NuvroLocation
                        ↓
                 Feature modules
                        ↓
                    Nuvro app
```

A feature module may depend on `NuvroModels`, `NuvroDomain`, `NuvroDesignSystem`,
`NuvroLocalization`, and `NuvroLocation` where genuinely required. Feature modules do not
depend on each other.

Forbidden:

- A View performing an HTTP call or touching persistence directly.
- A feature module implementing Match Score, ranking or any canonical backend logic.
- A feature module inventing its own visual constants instead of using the design system.

## UI state

Modern Swift concurrency: `async`/`await`, structured concurrency, `@MainActor` for UI state,
the Observation framework where appropriate, and constructor injection. No global mutable
singletons.

A View renders state, emits user intent and composes reusable components. Networking and
business orchestration never live inside a View.

Target a View file at 200 lines or fewer. The absolute maximum is 400, enforced by
`scripts/check-loc.sh`. A complex screen is composed from smaller components, not written as
one large file.

## Design system

```
NuvroDesignSystem/
├── Tokens/         Typography/   Components/
├── Glass/          Motion/       Layout/
└── Accessibility/
```

Liquid Glass rules:

1. Use native Apple behaviour where it exists.
2. Glass is for controls and functional layers, not for every card.
3. Content stays the visual focus.
4. Version-specific glass behaviour is encapsulated inside the design system.
5. OS availability checks never spread through feature screens.
6. Reduce Motion and accessibility preferences are respected.

Brand constants come from the design system. A feature that hardcodes a colour or a spacing
value is a defect.

## Accessibility

Dynamic Type, VoiceOver, Reduce Motion, sufficient contrast, meaningful accessibility labels
and adequate touch targets. The 2.5D onboarding has a Reduce Motion path, and its story is
understandable without parallax or large camera moves.

## Creating the Xcode project

When interface work starts:

1. Create the app project in Xcode under `App/`, product name `Nuvro`.
2. Add `Packages/NuvroKit` as a local package dependency.
3. Keep application code in `App/Nuvro/Application` and assets in `App/Nuvro/Resources`.
4. Everything reusable goes in a NuvroKit module, not in the app target.
5. The `ios` workflow builds the project automatically once it is committed.

## Running against a backend

The app talks to a backend only when it is told where one is. Nothing is configured by
default, so the app runs on prototype content and the prototype sign-in.

Set `INFOPLIST_KEY_NuvroAPIBaseURL` in `App/project.yml` for a real build, or pass the
launch argument `-apiBaseURL http://127.0.0.1:8080` to point a Debug build at a local
backend (only `https`, or `http` to loopback, is accepted). In Debug builds
`-devCredential <value>` signs in with the backend's `NUVRO_DEV_AUTH_CREDENTIAL`, which is
how the simulator and the UI tests exercise the real session flow without an Apple ID.

The client is generated at build time from `Packages/NuvroKit/Sources/NuvroAPI/openapi.json`.
Run `make ios-contract-sync` after changing the released contract; `make verify` fails when
that file is stale.
