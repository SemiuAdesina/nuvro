# Design

Brand direction and design source for NUVRO. No binary brand assets are committed yet, and
none are generated to fill a directory.

## Brand

```
Name:            NUVRO
Tone:            premium, global, simple, modern
Tagline:         The world, made personal.
Product phrase:  Pick a place. We handle the rest.
```

## Visual direction

Working direction, not approved final values:

```
deep sky blue          soft teal and aqua      soft warm gold
cloud white            twilight blue           deep navy
dawn and golden atmosphere
premium 2.5D travel artwork
```

**No final hex values are recorded here or hardcoded anywhere in the clients.** They are
fixed when the design system is formally approved, and they enter the codebase as design
system tokens, never as literals in a feature.

## Structure

```
design/
├── brand/
│   ├── logo/         reserved
│   └── app-icon/     reserved
├── onboarding/
│   ├── artwork/      reserved, layered 2.5D source and exports
│   └── references/   reserved
├── tokens/           reserved, the source that becomes platform tokens
└── README.md
```

## Onboarding artwork

The four-screen story is in `docs/product/onboarding.md`. The rule that matters for this
directory is how artwork enters the app:

- Optimised platform assets, not hundreds of hand-placed shapes inside a View.
- Art, motion behaviour, layout composition, accessibility fallback and screen state are five
  separate concerns in five separate places.
- The Reduce Motion path is designed, not bolted on.
- Large design source files are evaluated against repository size and Git LFS before being
  committed. A multi-hundred-megabyte source file in Git history is permanent.

## Tokens

`tokens/` will hold the single source that generates both platform token sets, so that iOS
and Android cannot drift on colour, spacing, radius, elevation and typography. Until the
design system is approved there is nothing to generate.
