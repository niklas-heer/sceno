package main

import (
	"fmt"
	"strings"

	"dagger/sceno/internal/dagger"
)

const (
	goImage   = "golang:1.23-bookworm"
	goVersion = "1.23"
)

var buildPlatforms = []struct {
	GOOS, GOARCH, Name string
}{
	{"darwin", "arm64", "sceno-darwin-arm64"},
	{"darwin", "amd64", "sceno-darwin-amd64"},
	{"linux", "amd64", "sceno-linux-amd64"},
	{"linux", "arm64", "sceno-linux-arm64"},
}

var releaseArchives = []struct {
	GOOS, GOARCH, Archive string
}{
	{"darwin", "arm64", "sceno_darwin_arm64.tar.gz"},
	{"darwin", "amd64", "sceno_darwin_amd64.tar.gz"},
	{"linux", "amd64", "sceno_linux_amd64.tar.gz"},
	{"linux", "arm64", "sceno_linux_arm64.tar.gz"},
}

func repoSource(source *dagger.Directory) *dagger.Directory {
	return source.
		WithoutDirectory(".git").
		WithoutDirectory("dist").
		WithoutDirectory("output")
}

func goTest(source *dagger.Directory) *dagger.Container {
	return dag.Container().
		From(goImage).
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "--no-install-recommends", "gcc", "libc6-dev"}).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("sceno-go-mod")).
		WithMountedCache("/go/cache", dag.CacheVolume("sceno-go-build")).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithEnvVariable("GOCACHE", "/go/cache").
		WithEnvVariable("CGO_ENABLED", "1").
		WithDirectory("/src", repoSource(source)).
		WithWorkdir("/src").
		WithExec([]string{"go", "mod", "download"})
}

func goBase(source *dagger.Directory) *dagger.Container {
	return dag.Container().
		From(goImage).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("sceno-go-mod")).
		WithMountedCache("/go/cache", dag.CacheVolume("sceno-go-build")).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithEnvVariable("GOCACHE", "/go/cache").
		WithEnvVariable("CGO_ENABLED", "0").
		WithDirectory("/src", repoSource(source)).
		WithWorkdir("/src").
		WithExec([]string{"go", "mod", "download"})
}

func ldflags(version, commit, builtAt string) string {
	commit = strings.TrimSpace(commit)
	if commit == "" {
		commit = "unknown"
	}
	if len(commit) > 7 {
		commit = commit[:7]
	}
	builtAt = strings.TrimSpace(builtAt)
	if builtAt == "" {
		builtAt = "unknown"
	}
	version = strings.TrimSpace(version)
	if version == "" {
		version = "dev"
	}
	// version may be a bash expression like ${VERSION}
	return fmt.Sprintf("-s -w -X github.com/niklas-heer/sceno/internal/version.Version=%s -X github.com/niklas-heer/sceno/internal/version.Commit=%s -X github.com/niklas-heer/sceno/internal/version.Date=%s",
		version, commit, builtAt)
}

func readVersionScript() string {
	return `VERSION=$(tr -d '[:space:]' < internal/version/VERSION)`
}

func smokeScript() string {
	return `
set -euo pipefail
./sceno version
./sceno version --json | grep -q '"version"'
./sceno docs --json | grep -q '"start_here"'
./sceno docs guide --json | grep -q '"tool": "sceno"'
./sceno docs errors --json | grep -q 'missing_node'
./sceno goals | head -n 5 >/dev/null
for f in examples/*.kdl; do
  ./sceno validate -i "$f" --json | grep -q '"ok": true' || { echo "validate failed: $f"; exit 1; }
done
./sceno render -i examples/self-service.kdl -o /tmp/sceno-smoke --all
test -s /tmp/sceno-smoke.svg
test -s /tmp/sceno-smoke.png
test -s /tmp/sceno-smoke.pdf
test -s /tmp/sceno-smoke.html
test -s /tmp/sceno-smoke.slides.html
echo "smoke ok"
`
}

func lintScripts(source *dagger.Directory) *dagger.Container {
	return dag.Container().
		From("alpine:3.20").
		WithExec([]string{"apk", "add", "--no-cache", "bash"}).
		WithDirectory("/src", repoSource(source)).
		WithWorkdir("/src").
		WithExec([]string{"bash", "-n", "scripts/install.sh"}).
		WithExec([]string{"bash", "-n", "scripts/bump-version.sh"})
}
