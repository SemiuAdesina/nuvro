# Disabled GitHub automation

These workflow definitions are intentionally stored outside `.github/workflows`,
so GitHub Actions cannot discover or run them.

To re-enable a workflow later, review it first and move its YAML file back into
`.github/workflows/`.

Dependabot is also disabled in `.github/dependabot.yml.disabled`. Rename that
file back to `.github/dependabot.yml` only when automated dependency pull
requests are wanted.
