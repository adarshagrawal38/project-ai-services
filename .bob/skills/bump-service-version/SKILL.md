---
name: bump-service-version
description: Use when the user wants to bump, update, or change the version/image tag of any service, UI component, or the catalog backend (ai-services) inside the ai-services project. Runs make release to auto-detect all modified components, bump their Makefile TAGs, and sync all values.yaml files in one shot.
metadata:
  argument-hint: "<optional: base-branch>"
---

# Bump Service Version

Bumps the TAG in every modified component Makefile and syncs all corresponding `values.yaml`
files using the repo-root `make release` target. No manual component selection required —
changed components are auto-detected via git diff.

## Available make targets

| Target | What it does |
|---|---|
| `make release` | Auto-detect changed components → bump TAGs → sync all `values.yaml` |
| `make release REGISTRY=<url>` | Same, with a custom registry |
| `make release BASE_BRANCH=<ref>` | Same, diffing against a different base branch |

---

## Step 1 — Run release

```
execute_command: make release
```

This runs `bump-tags` (auto-detects changed components from git diff, bumps each Makefile TAG
from the base-branch value + 1) followed by `update-tags` (syncs every modified Makefile's
new TAG into the corresponding `values.yaml` files).

---

## Step 2 — Verify

Grep to confirm no stale image tags remain:

```
grep pattern: TAG\?=
path: ai-services/Makefile
```

If any unexpected stale references remain outside `ai-services/assets/applications/`,
investigate and fix before reporting.

---

## Step 3 — Report

Summarise what was changed: list each component whose Makefile TAG was bumped and each
`values.yaml` that was updated, in a single table.

| File | Change |
|------|--------|
| `<component>/Makefile` | `TAG=<old>` → `TAG=<new>` |
| `<values.yaml path>` | `<image>:<old>` → `<image>:<new>` |

Confirm no stale references were found.
