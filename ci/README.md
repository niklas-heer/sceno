# Sceno CI (Dagger)

Pipeline-as-code for Sceno. Same commands run locally and in GitHub Actions.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) (or Colima)
- [Dagger CLI](https://docs.dagger.io/install) v0.20.8+

## Commands

```bash
# From repo root
make ci                              # full CI
make ci-test                         # go test -race
make ci-smoke                        # build + integration smoke

dagger functions                     # list pipeline functions
dagger call test --source=.
dagger call ci --source=. --commit=$(git rev-parse HEAD)
dagger call release --source=. --tag=v0.1.0 export --path=dist
```

## Functions

| Function | Description |
|----------|-------------|
| `test` | `go mod verify` + `go test -race` |
| `smoke` | Build CLI + validate examples + render smoke exports |
| `lint-scripts` | `bash -n` on install scripts |
| `build` | Cross-compile one platform |
| `build-all` | All four platform binaries |
| `ci` | Full pipeline (test → smoke → lint → build-all) |
| `release` | Verify tag/VERSION, test, build release tarballs + SHA256SUMS |

## GitHub Actions

`.github/workflows/ci.yml` and `release.yml` are thin wrappers around `dagger call`.
