# Sceno CI (Dagger)

Pipeline-as-code for Sceno. Same commands run locally and in GitHub Actions.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) (or Colima)
- [Dagger CLI](https://docs.dagger.io/install) v0.20.8+

The pipeline uses Go **1.27.1**. Its test container installs Node.js for the dependency-free preview state tests; the distributed Sceno binary embeds all browser assets and does not require Node.js. macOS binaries require macOS 13 or later.

## Commands

```bash
# From repo root
mask ci                              # full CI
mask ci-test                         # preview UI state tests + go test -race
mask ci-smoke                        # build + integration smoke

dagger functions                     # list pipeline functions
dagger call test --source=.
dagger call ci --source=. --commit=$(git rev-parse HEAD)
dagger call release --source=. --tag=v0.3.0 export --path=dist
```

## Functions

| Function | Description |
|----------|-------------|
| `test` | Preview UI state tests + `go mod verify` + `go test -race` |
| `smoke` | Build CLI + validate, advise, describe, and export all examples and starter templates |
| `lint-scripts` | `bash -n` on install scripts |
| `build` | Cross-compile one platform |
| `build-all` | All four platform binaries |
| `ci` | Full pipeline (test → smoke → lint → build-all) |
| `release` | Verify tag/VERSION, test, build release tarballs + SHA256SUMS |

## GitHub Actions

`.github/workflows/ci.yml` and `release.yml` are thin wrappers around `dagger call`.
