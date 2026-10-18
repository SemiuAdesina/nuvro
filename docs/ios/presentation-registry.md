# iOS presentation registry

Every sheet, cover, alert, confirmation dialog, picker, permission explainer and transient
banner in the iOS app. A new presentation is added here in the same change that introduces it,
so modal behaviour is never duplicated or forgotten.

Legend: preserved means the owning model keeps its state across dismissal.

## Sheets

| Owner | Trigger | Presentation | Dismissal | Preserved | Destructive |
| --- | --- | --- | --- | --- | --- |
| SETUP-02/03/04 | Choose or add a country | `SearchablePickerSheet` over Foundation ISO regions with flags, large | Select | Yes | No |
| SETUP-05 | Choose airport | `SearchablePickerSheet<Airport>` | Select | Yes | No |
| SETUP-06 | Choose currency | `SearchablePickerSheet` over Foundation ISO currencies | Select | Yes | No |
| 29 Profile | Travel access row | `TravelSetupView` in edit mode for one step, large | Done saves, Cancel discards | Yes | No |
| 06 Match Setup | Change departure | `SearchablePickerSheet` (airports), large detent | Select or Done | Yes | No |
| 06 Match Setup | Tap dates | `DateSelectionSheet`, medium/large | Done commits | Yes | No |
| 06 Match Setup | Change travellers | `CountPickerSheet`, medium | Done | Yes | No |
| 09 Trip Plan Hub | Edit trip menu | `DateSelectionSheet`, `CountPickerSheet`, `TripBudgetSheet` | Done applies edits | Yes | No |
| 11 Budget | Adjust budget | `TripBudgetSheet` on the same `TripPlanModel` | Done applies edits | Yes | No |
| SETUP-03/04 | Change on a passport or permit card | Same `SearchablePickerSheet`; the pick replaces that card, dismissing keeps it | Select | Yes | No |
| 10 Full Itinerary | Move activity | `OptionPickerSheet` of other days | Select | Yes | No |
| 16 Explore Nearby | First open, location undetermined or denied | Permission sheet, medium | Continue, Not now, Use destination centre, Open Settings | Yes | No |
| 15 Language Companion | Tap microphone / choose Camera | Permission sheet, medium | Continue, Not now, Type instead, Open Settings | Yes | No |
| 18 Smart Alerts | Alert preferences | `AlertPreferencesSheet`, medium/large | Done | Yes | No |
| 23 Flights | View flight | `PartnerHandoffSheet`, medium; Continue disabled until a live booking link exists | Cancel or Continue | Yes | No |
| 16 Explore Nearby | Go | Hands off to Apple Maps walking directions via `openURL` | System | n/a | No |
| 15 Language Companion | Play / Copy | `AVSpeechSynthesizer` in the target language; pasteboard plus "Copied" transient banner | Auto | n/a | No |
| 17 Tickets & Documents | Add to Wallet / Add all to Wallet | Transient banner stating passes need the live booking service (no `.pkpass` exists yet) | Auto | n/a | No |
| 14 Travel Mode | Next card Navigate / Ticket | Push to Explore Nearby / Tickets & Documents (same as the quick-action tiles) | Back | n/a | No |
| 27 Group Planning | Invite traveller | `InviteTravellerSheet`, medium | Cancel or Send | Email cleared on send | No |
| 29 Profile | "What NUVRO has learned" card | `LearningPreferenceCard`: decorative `profile-learning-attendant` illustration between the label and the switch, clipped to the card, hidden from VoiceOver, dropped at accessibility text sizes; one-time fade-in honours Reduce Motion | n/a | n/a | No |
| 29 Profile | Tap a preference row (title and value stay on one line; multi-value rows read "Beach, Food +4") | `OptionPickerSheet` or `MultiOptionPickerSheet` | Select or Done | Yes | No |
| 30 Language & region | App language, Region, Currency, Home airport | `SearchablePickerSheet`, large | Select or Done | Yes | No |
| 30 Language & region | Languages I speak | `MultiOptionPickerSheet`, minimum one | Done | Yes | No |
| 31 Settings | Delete account row | `AccountDeletionSheet`, large, non-dismissable while deleting | Cancel | Failure shown inline | Yes |

## Full-screen presentations

| Owner | Trigger | Presentation | Dismissal |
| --- | --- | --- | --- |
| 17 Tickets & Documents | Show ticket | `fullScreenCover` scannable ticket | Tap anywhere |
| 15 Language Companion | Show to someone | In-place overlay large translation | Tap to return |
| Root | No session after onboarding | `AuthenticationView` replaces the tab UI | Session established |

## Alerts

| Owner | Trigger | Title | Actions |
| --- | --- | --- | --- |
| 10 Full Itinerary | Remove a reserved, booked or ticketed activity | Remove this reservation? | Remove (destructive), Cancel |
| 17 Tickets & Documents | Remove download | Remove the offline copy? | Remove download (destructive), Cancel |
| 20 Trips Library | Delete trip | Delete this trip? | Delete trip (destructive), Cancel |
| 27 Group Planning | Remove traveller | Remove from this trip? | Remove traveller (destructive), Cancel |
| 27 Group Planning | Leave trip | Leave this trip? | Leave trip (destructive), Cancel |
| 31 Settings | Delete account, second step | Delete your account? | Delete account (destructive), Cancel |

## Confirmation dialogs

| Owner | Trigger | Options |
| --- | --- | --- |
| 09 Trip Plan Hub | Tap the trip identity header | Change dates, Change travellers, Change budget, Travelling together, Cancel |
| 20 Trips Library | Ellipsis or long press on a trip | Download for offline, Archive trip, Delete trip (destructive), Cancel |
| 27 Group Planning | Ellipsis on a non-owner | Change role to Editor/Viewer, Remove traveller (destructive), Cancel |
| 31 Settings | Sign out | Sign out (destructive), Cancel |

## Context menus

| Owner | Element | Actions |
| --- | --- | --- |
| 10 Full Itinerary | Activity ellipsis menu | Move, Remove |
| 17 Tickets & Documents | Document row | Download for offline or Remove download |
| 20 Trips Library | Trip card | Same as the dialog |
| 26 Packing | Checklist item | Remove (destructive, reversible by re-adding, so no alert) |

## Pickers

| Picker | Component | Used by |
| --- | --- | --- |
| Departure and home airport | `SearchablePickerSheet<Airport>` | 06, 30 |
| Travel dates, exact or flexible | `DateSelectionSheet` | 06 |
| Travellers | `CountPickerSheet` | 06 |
| App language | `SearchablePickerSheet<Language>` | 30 |
| Spoken languages | `MultiOptionPickerSheet` | 30 |
| Region, currency | `SearchablePickerSheet<PickerOption>` | 30 |
| Interests, accessibility | `MultiOptionPickerSheet` | 29 |
| Pace, weather, crowds, flights, stays, budget style | `OptionPickerSheet` | 29 |
| Itinerary day | `OptionPickerSheet` | 10 |
| Units, time and date format | `SegmentedChoice` inline | 30 |
| Notification preferences | `AlertPreferencesSheet` toggles | 18, 31 |

## Permission explainers

| Permission | Where it is asked | Before the system prompt | If denied |
| --- | --- | --- | --- |
| Location | 16 Explore Nearby on open | "Find what's around you" sheet | Open Settings or use destination centre |
| Microphone | 15 Speak mode, on tap | "Speak to translate" sheet | Open Settings or type instead |
| Camera | 15 Camera mode | "Read menus and signs" sheet | Inline card, type instead |
| Notifications | 18 Smart Alerts inline card | "Stay ahead of changes" card | Inline card with Open Settings |

Permissions are never requested at launch or during onboarding. 31 Settings shows live states
and links to iOS Settings.

## Transient banners

One `TransientFeedback` instance lives at the root; showing a message replaces the current one,
nothing stacks. Triggers: Saved / Removed from saved (05), Preference updated (29), Added to today
(16), Downloaded for offline and Offline copy removed (17, 20), Trip archived / deleted (20),
Traveller removed and Invitation sent (27), Removed from Day N and Moved to Day N (10).

The offline banner is a separate root overlay driven by `ConnectivityMonitor`, with online,
offline and reconnecting states.
