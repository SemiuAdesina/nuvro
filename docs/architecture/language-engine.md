# Language engine

E10 helps a traveller communicate. It is not a chatbot.

## Canonical languages

Languages are BCP 47 tags: `en`, `pt-PT`, `pt-BR`. Display names such as "Portuguese" are
rejected. Parsing normalises case, so `pt-pt` becomes `pt-PT`, and regional variants compare
as the same language when that matters.

## Resolving the pair

1. An explicitly requested target wins.
2. Otherwise the destination's country maps to a language through an explicit table.
3. The source is the traveller's app language, defaulting to English.

Residence is never used to infer language. An English app in Portugal translates to `pt-PT`; a
French app in Portugal still translates from `fr-FR`. An unmapped region produces no invented
target.

## Translation

The engine validates before the provider is called: the target must parse, the text must be
non-empty and under five thousand characters, and context is truncated to two hundred
characters. If the provider claims to have detected the source but returns nothing usable, the
result is rejected rather than reported as a translation.

No translation provider is configured. The port exists, the engine reports itself
unconfigured, and the endpoint returns an unavailable error. Nothing is faked.

## Phrases

A small curated catalog covers hotel, restaurant, transport, directions, shopping, emergency,
medical and general use. It is deliberately about two dozen phrases, not five hundred. When a
provider exists each phrase is translated; without one, phrases are returned untranslated and
explicitly marked as such.

Emergency and medical phrases are literal communication only: "I need help", "Please call
emergency services", "I am allergic to this". They contain no dosage, diagnosis or treatment
advice, and a test enforces that.

## Camera and speech

Optical character recognition and speech recognition stay on the device. Apple's native APIs
extract text; the backend translates that text. NUVRO does not receive photographs or
continuous microphone audio, which keeps a large privacy surface off the server entirely.

## Privacy

Translated text can contain anything a traveller needs to say, including private detail. It is
processed transiently. Nothing is persisted by default, and no source or translated text is
logged or attached to telemetry.
