# Localization

NUVRO is global from the first screen, not localised later.

## Language is not locale

These are separate concepts and are stored separately:

| Concept | Example | Determines |
| --- | --- | --- |
| Language | `fr` | The words the traveller reads |
| Locale | `fr-CA` | Number, date and currency formatting conventions |
| Currency | `EUR` | The currency amounts are presented in |
| Home country | `NG` | Departure defaults, tax and pricing context |
| Departure airport | `LHR` | Flight search origin |
| Timezone | `Europe/London` | The traveller's current local time |
| Units | metric | Distance and temperature presentation |
| Passport and nationality | `NG` | Entry requirement eligibility |

A traveller can differ on every one of these at once. No field is inferred from another.

## Device first, traveller always

The app detects the device language on first launch and lets the traveller change it during
onboarding and afterwards from their profile. The chosen language is part of the profile and
syncs across devices.

## Client strings

UI strings are never hardcoded across Views or Composables.

- iOS keeps every user-facing string in `NuvroLocalization`, resolved by `Language` through typed
  copy structs (`AuthenticationCopy`, `StateCopy`, `PickerCopy`, `PermissionCopy` and the screen
  copies). Feature modules never hold a literal. Moving the values into string catalogues is the
  next step; the resolution boundary does not change.
- Android uses Android resource localization in `:core:localization` and per-feature
  resources.

Neither client invents its own string table format.

## Backend

Canonical business models hold no English UI copy. A destination model holds identifiers,
facts and structured values, not a display sentence.

Requests carry language and locale context where it changes the result. `Accept-Language` is
declared in the OpenAPI contract as a standard header. AI requests carry language and locale
so that a generated explanation is written in the traveller's language rather than translated
afterwards.

Canonical units are stored and returned. Clients convert:

```
kilometres and miles        Celsius and Fahrenheit
12-hour and 24-hour         currency formatting        date formatting
```

## Time

Backend timestamps are UTC RFC3339. Destination local times keep an IANA timezone identifier
alongside the value. A local time without a zone is ambiguous and is never stored or returned.

## Money

Amounts cross the contract as integer minor units plus an ISO 4217 code. Formatting is a
client concern and depends on locale, not language. `NuvroModels.Money` already implements
this, including zero-decimal currencies such as JPY, and is tested against the contract shape.

## Right to left and text growth

Layouts accommodate right-to-left languages and text that grows on translation. A component
that only works with English-length strings is not finished.
