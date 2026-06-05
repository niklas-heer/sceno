// Sceno CI/CD pipeline — run locally with: dagger call ci --source=.
package main

import (
	"context"
	"fmt"
	"strings"

	"dagger/sceno/internal/dagger"
)

type Sceno struct{}

// Test runs go mod verify and the full test suite with race detector.
func (m *Sceno) Test(
	ctx context.Context,
	// +optional
	// +defaultPath="/"
	source *dagger.Directory,
) (string, error) {
	ctr := goTest(source).
		WithExec([]string{"go", "mod", "verify"}).
		WithExec([]string{"go", "test", "-race", "-count=1", "./..."})
	return ctr.Stdout(ctx)
}

// Smoke builds the CLI and runs integration smoke checks.
func (m *Sceno) Smoke(
	ctx context.Context,
	// +optional
	// +defaultPath="/"
	source *dagger.Directory,
	// Git commit SHA (shortened in ldflags)
	// +optional
	commit string,
	// Build timestamp (RFC3339)
	// +optional
	builtAt string,
) (string, error) {
	ctr := goBase(source).
		WithExec([]string{"bash", "-ec", readVersionScript() + `
go build -ldflags "` + ldflags("${VERSION}", commit, builtAt) + `" -o sceno ./cmd/sceno
` + smokeScript()})
	return ctr.Stdout(ctx)
}

// LintScripts syntax-checks shell install scripts.
func (m *Sceno) LintScripts(
	ctx context.Context,
	// +optional
	// +defaultPath="/"
	source *dagger.Directory,
) (string, error) {
	return lintScripts(source).Stdout(ctx)
}

// Build compiles sceno for one GOOS/GOARCH pair.
func (m *Sceno) Build(
	ctx context.Context,
	// +optional
	// +defaultPath="/"
	source *dagger.Directory,
	// +optional
	// +default="linux"
	goos string,
	// +optional
	// +default="amd64"
	goarch string,
	// +optional
	commit string,
	// +optional
	builtAt string,
) (*dagger.File, error) {
	name := fmt.Sprintf("sceno-%s-%s", goos, goarch)
	ctr := goBase(source).
		WithEnvVariable("GOOS", goos).
		WithEnvVariable("GOARCH", goarch).
		WithExec([]string{"bash", "-ec", readVersionScript() + fmt.Sprintf(`
go build -ldflags "%s" -o %s ./cmd/sceno
test -s %s
echo "built %s"
`, ldflags("${VERSION}", commit, builtAt), name, name, name)})
	if _, err := ctr.Stdout(ctx); err != nil {
		return nil, err
	}
	return ctr.File(name), nil
}

// BuildAll cross-compiles sceno for macOS and Linux (arm64 + amd64).
func (m *Sceno) BuildAll(
	ctx context.Context,
	// +optional
	// +defaultPath="/"
	source *dagger.Directory,
	// +optional
	commit string,
	// +optional
	builtAt string,
) (*dagger.Directory, error) {
	dir := dag.Directory()
	for _, p := range buildPlatforms {
		file, err := m.Build(ctx, source, p.GOOS, p.GOARCH, commit, builtAt)
		if err != nil {
			return nil, fmt.Errorf("build %s/%s: %w", p.GOOS, p.GOARCH, err)
		}
		dir = dir.WithFile(p.Name, file)
	}
	return dir, nil
}

// Ci runs the full continuous integration pipeline (test, smoke, scripts, cross-build).
func (m *Sceno) Ci(
	ctx context.Context,
	// +optional
	// +defaultPath="/"
	source *dagger.Directory,
	// +optional
	commit string,
	// +optional
	builtAt string,
) (string, error) {
	var b strings.Builder

	testOut, err := m.Test(ctx, source)
	if err != nil {
		return "", fmt.Errorf("test: %w", err)
	}
	b.WriteString("== test ==\n")
	b.WriteString(testOut)

	smokeOut, err := m.Smoke(ctx, source, commit, builtAt)
	if err != nil {
		return "", fmt.Errorf("smoke: %w", err)
	}
	b.WriteString("\n== smoke ==\n")
	b.WriteString(smokeOut)

	lintOut, err := m.LintScripts(ctx, source)
	if err != nil {
		return "", fmt.Errorf("lint-scripts: %w", err)
	}
	b.WriteString("\n== lint-scripts ==\n")
	b.WriteString(lintOut)

	if _, err := m.BuildAll(ctx, source, commit, builtAt); err != nil {
		return "", fmt.Errorf("build-all: %w", err)
	}
	b.WriteString("\n== build-all ==\nall platform binaries built\n")
	b.WriteString("\nci ok\n")
	return b.String(), nil
}

// Release verifies tag matches VERSION, runs tests, and produces release tarballs.
func (m *Sceno) Release(
	ctx context.Context,
	// +optional
	// +defaultPath="/"
	source *dagger.Directory,
	// Release tag, e.g. v0.1.0
	// +required
	tag string,
	// +optional
	commit string,
	// +optional
	builtAt string,
) (*dagger.Directory, error) {
	tag = strings.TrimPrefix(strings.TrimSpace(tag), "v")
	if tag == "" {
		return nil, fmt.Errorf("tag is required (e.g. v0.1.0)")
	}

	verify := goBase(source).
		WithExec([]string{"bash", "-ec", fmt.Sprintf(`
set -euo pipefail
FILE=$(tr -d '[:space:]' < internal/version/VERSION)
if [ "%s" != "$FILE" ]; then
  echo "tag v%s does not match VERSION file ($FILE)" >&2
  exit 1
fi
echo "release version: %s"
`, tag, tag, tag)})
	if _, err := verify.Stdout(ctx); err != nil {
		return nil, err
	}

	if _, err := m.Test(ctx, source); err != nil {
		return nil, fmt.Errorf("test: %w", err)
	}

	out := dag.Directory()
	for _, p := range releaseArchives {
		archive, err := m.releaseArchive(ctx, source, p.GOOS, p.GOARCH, p.Archive, tag, commit, builtAt)
		if err != nil {
			return nil, err
		}
		out = out.WithFile(p.Archive, archive)
	}

	sumCtr := dag.Container().
		From("alpine:3.20").
		WithDirectory("/dist", out).
		WithWorkdir("/dist").
		WithExec([]string{"sh", "-c", "sha256sum *.tar.gz > SHA256SUMS && cat SHA256SUMS"})
	sums := sumCtr.File("/dist/SHA256SUMS")
	install := repoSource(source).File("scripts/install.sh")
	return out.WithFile("SHA256SUMS", sums).WithFile("install.sh", install), nil
}

func (m *Sceno) releaseArchive(
	ctx context.Context,
	source *dagger.Directory,
	goos, goarch, archive, version, commit, builtAt string,
) (*dagger.File, error) {
	ctr := goBase(source).
		WithEnvVariable("GOOS", goos).
		WithEnvVariable("GOARCH", goarch).
		WithExec([]string{"bash", "-ec", fmt.Sprintf(`
set -euo pipefail
go build -ldflags "%s" -o sceno ./cmd/sceno
test -s sceno
cp scripts/install.sh install.sh
tar -czf %s sceno LICENSE README.md install.sh
`, ldflags(version, commit, builtAt), archive)})
	if _, err := ctr.Stdout(ctx); err != nil {
		return nil, err
	}
	return ctr.File(archive), nil
}
