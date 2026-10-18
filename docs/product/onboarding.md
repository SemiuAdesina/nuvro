# Onboarding

Onboarding is a four-screen story that ends by handing the traveller to Discover. It lives in
one iOS module, `NuvroOnboarding`, and one Android feature module, `:feature:onboarding`.

## Scope

The module owns exactly this and nothing else:

```
Splash              Onboarding 1        Onboarding 2        Onboarding 3
Language selector   Skip                Progress indicator
Reduce Motion behaviour                 Discover handoff
```

Anything beyond this list belongs in another module.

## The story

| Screen | Copy |
| --- | --- |
| S0 | **NUVRO** — The world, made personal. |
| O1 | **Not sure where to go?** — Your budget. Your dates. Your vibe. |
| O2 | **Pick a place. We handle the rest.** |
| O3 | **Your trip doesn't stop when you arrive.** |

Final actions:

```
Find my vacation          I already have a trip
```

The first goes to Discover. The second goes to Plan.

## Language

The app detects the device language on first launch. The language selector lets the traveller
change it during onboarding, before any content depends on it. Language is not locale, and
choosing a language does not choose a currency or a home country.

## Status

Implemented in `NuvroOnboarding` from the approved Claude Design canvas, verified on an
iPhone 17 Pro simulator against the design's 402 by 874 coordinates. The geometry, camera
poses, timings and copy are the design's own values; see `clients/ios/README.md` for how they
scale to other devices and which stand-ins remain.

## Visual direction

Premium 2.5D travel artwork with subtle motion, built as three parallax layers under one
camera. The architectural requirement is how it is built, not what it looks like:

- Art is an optimised platform asset, never hundreds of hand-placed shapes inside a View.
- Motion behaviour is separate from the art and from the layout.
- Layout composition, motion, accessibility fallback and screen state are four separate
  concerns in four separate files.
- Large source design files are evaluated against repository size and Git LFS before being
  committed.

## Accessibility

The 2.5D presentation has a Reduce Motion path. Parallax and large camera moves are
decoration; the story must be fully understandable without them.

Dynamic Type, VoiceOver, sufficient contrast, meaningful accessibility labels and adequate
touch targets apply here as everywhere. A progress indicator announces position in the flow.
Skip is reachable at every step and is never the only way out.

Android provides the equivalent through native accessibility behaviour rather than by copying
the iOS implementation.
