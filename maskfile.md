# Sceno tasks

Development and release commands for the Sceno CLI.

Requires [mask](https://github.com/jacobdeichert/mask) (`brew install mask`).

Run `mask --help` for the full list. Common commands:

| Task | Command |
|------|---------|
| Build | `mask build` |
| Test | `mask test` |
| Quick smoke | `mask verify` |
| Full CI | `mask ci` |
| Release | `mask release` |
| Cross-compile | `mask dist` |

Legacy `make` targets map 1:1 — e.g. `make ci` → `mask ci`, `make bump-patch` → `mask bump-patch`.

## build

> Build the sceno binary with embedded version metadata

```sh
source scripts/lib.sh
sceno_env
go build -ldflags="$LDFLAGS" -o "$BINARY" "$CMD"
echo "built ./$BINARY ($VERSION)"
```

## test

> Run unit tests with race detector

```sh
go test -race -count=1 ./...
```

## install

> Install sceno to GOPATH/bin

```sh
source scripts/lib.sh
sceno_env
go install -ldflags="$LDFLAGS" "$CMD"
```

## clean

> Remove build artifacts

```sh
rm -f sceno
rm -rf dist/
```

## lint

> Run go vet and tests

```sh
go vet ./...
go test ./...
```

## examples

> Run pipeline example tests

```sh
go test ./internal/pipeline/ -run Examples -count=1
```

## verify

> Quick local smoke test (build, validate, render)

```sh
source scripts/lib.sh
sceno_env
go build -ldflags="$LDFLAGS" -o "$BINARY" "$CMD"
./"$BINARY" version
./"$BINARY" validate -i examples/self-service.kdl --json | grep -q '"ok": true'
./"$BINARY" render -i examples/self-service.kdl -o dist/smoke --all
for f in dist/smoke.*; do test -s "$f" || (echo "missing $f" && exit 1); done
echo "verify ok ($VERSION)"
```

## ci

> Full CI via Dagger (same as GitHub Actions)

```sh
source scripts/lib.sh
sceno_env
dagger call ci --source=. --commit="$COMMIT" --built-at="$DATE"
```

## ci-test

> Dagger: go test -race only

```sh
dagger call test --source=.
```

## ci-smoke

> Dagger: build + integration smoke checks

```sh
source scripts/lib.sh
sceno_env
dagger call smoke --source=. --commit="$COMMIT" --built-at="$DATE"
```

## ci-release

> Dagger: dry-run release artifacts to dist/

```sh
source scripts/lib.sh
sceno_env
dagger call release --source=. --tag="v$VERSION" --commit="$COMMIT" --built-at="$DATE" export --path=dist
ls dist/*.tar.gz dist/SHA256SUMS dist/install.sh
```

## dist

> Cross-compile release binaries for all platforms into dist/

```sh
source scripts/lib.sh
sceno_env
mkdir -p dist
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o dist/sceno-darwin-arm64 "$CMD"
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o dist/sceno-darwin-amd64 "$CMD"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o dist/sceno-linux-amd64 "$CMD"
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o dist/sceno-linux-arm64 "$CMD"
echo "dist ok ($(VERSION))"
ls -la dist/sceno-*
```

## darwin-arm64

> Build macOS Apple Silicon binary into dist/

```sh
source scripts/lib.sh
sceno_env
mkdir -p dist
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o dist/sceno-darwin-arm64 "$CMD"
echo "built dist/sceno-darwin-arm64 ($VERSION)"
```

## bump-patch

> Manually bump patch version in internal/version/VERSION

Prefer `mask release` for automated semver. Use this only to override manually.

```sh
./scripts/bump-version.sh patch
```

## bump-minor

> Manually bump minor version in internal/version/VERSION

```sh
./scripts/bump-version.sh minor
```

## bump-major

> Manually bump major version in internal/version/VERSION

```sh
./scripts/bump-version.sh major
```

## release-tag

> Create an annotated git tag for the current VERSION (does not push)

Prefer `mask release` for the full flow. Use after a manual VERSION bump.

```sh
source scripts/lib.sh
sceno_env
test -n "$VERSION" || (echo "VERSION file missing" && exit 1)
git diff --quiet internal/version/VERSION || (echo "Commit VERSION changes first" && exit 1)
git tag -a "v$VERSION" -m "Release v$VERSION"
echo "Created tag v$VERSION"
echo "Push with: git push origin main && git push origin v$VERSION"
```

## release

> Suggest semver from conventional commits, run CI, bump VERSION, tag, and push

Analyzes commits since the last tag and suggests the next version (`feat:` → minor, `fix:` → patch on 0.x). Then runs the full release pipeline.

**OPTIONS**

* yes
    * flags: -y --yes
    * desc: Skip confirmation prompt
* dry_run
    * flags: -n --dry-run
    * desc: Preview suggested version without making changes
* skip_ci
    * flags: --skip-ci
    * desc: Skip the CI pipeline before releasing
* force
    * flags: -f --force
    * desc: Allow release from a branch other than main
* version
    * flags: -V --version
    * type: string
    * desc: Override the suggested version (e.g. 0.2.1)

```sh
./scripts/release.sh
```

## next-version

> Print the suggested next semver (no v prefix)

```sh
./scripts/next-version.sh next
```

## next-version-tag

> Print the suggested next semver tag (with v prefix)

```sh
./scripts/next-version.sh tag
```
