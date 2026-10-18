# Personalization data

This describes exactly what NUVRO stores about how a traveller behaves, what it refuses to
store, and what a traveller can do about it.

## What is stored

`personalization_signals` holds one row per observed action: a signal type, a subject type,
a subject value, a strength, an optional trip reference and a timestamp. A row says
"a coastal destination was saved", never "this traveller likes the sea".

`learned_preferences` holds one row per inference: a preference key, a confidence, a decayed
score, an evidence count, an explanation code, a status and a rules version.

`trip_feedback` holds one row per trip: typed aspects the traveller selected from a closed
list, a pace assessment, a travel burden assessment and a return intent.

## What is never inferred

Personalization must not use or derive race, ethnicity, religion, political ideology, sexual
orientation, medical condition, disability beyond an explicitly stated accessibility
requirement, or financial status beyond the budget the traveller set for a trip.

Passport, visa friction, residence and entry eligibility are feasibility inputs for E15 and
are never preference-learning signals. A traveller who is repeatedly refused a visa has not
expressed a taste.

Commission and partner revenue are never personalization inputs. A learned preference cannot
be created, strengthened or weighted by what a booking pays.

## Free text

Post-trip feedback is a fixed vocabulary of aspects. There is no free-text field, so there is
no prose about a traveller to store, mine or leak. This is a design constraint, not a
limitation waiting to be lifted.

## Logging and metrics

Raw feedback, personalization history and sensitive profile attributes are never logged.
User identifiers are never used as metric labels; personalization metrics are counted by
rules version, preference key and outcome only.

Insight text is never persisted. The API carries codes and parameters, and the client renders
the sentence.

## Traveller control

| Action | Endpoint | Effect |
| --- | --- | --- |
| See everything learned | `GET /v1/me/personalization/insights` | Every inference, with its confidence, evidence count and explanation |
| See how it is applied | `GET /v1/me/personalization` | Explicit and learned influence side by side, plus what is suppressed |
| Confirm an inference | `PATCH .../insights/{insightId}` | Promotes it to confirmed weight |
| Dismiss an inference | `PATCH .../insights/{insightId}` | Removes its influence permanently; re-learning cannot resurrect it |
| Suppress an inference | `PATCH .../insights/{insightId}` | Silences it for a window without deleting the evidence |
| Turn learning off | `PATCH /v1/me/personalization` | No historical behaviour informs any recommendation, immediately and reversibly |
| Clear learning | `DELETE /v1/me/personalization` | Deletes signals and learned preferences; keeps explicit preferences, trips and feedback |
| Revise feedback | `PUT /v1/trips/{tripId}/feedback` | Replaces the earlier submission |

Turning learning off is a read-time gate, so it takes effect on the next request with no
deletion and no waiting. Clearing is a deletion and is not reversible.

## Retention

Signals older than the aggregation window stop being read. Decay means old evidence loses
influence long before it stops being stored, so a stale taste fades rather than persisting
until a purge runs.

## Why explainability is a storage decision

Every learned row carries the explanation code and rules version that produced it. When the
rules change, existing inferences remain attributable to the version that created them, and
a traveller asking "why does NUVRO think this?" gets an answer derived from stored evidence
rather than reconstructed after the fact.
