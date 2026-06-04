.PHONY: build test install clean lint examples dist release-tag bump-patch bump-minor bump-major verify

BINARY := sceno
CMD := ./cmd/sceno
VERSION := $(shell tr -d '[:space:]' < internal/version/VERSION)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X github.com/niklas-heer/sceno/internal/version.Version=$(VERSION) \
	-X github.com/niklas-heer/sceno/internal/version.Commit=$(COMMIT) \
	-X github.com/niklas-heer/sceno/internal/version.Date=$(DATE)

build:
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) $(CMD)

test:
	go test -race -count=1 ./...

install:
	go install -ldflags="$(LDFLAGS)" $(CMD)

clean:
	rm -f $(BINARY)
	rm -rf dist/

lint:
	go vet ./...
	go test ./...

examples:
	go test ./internal/pipeline/ -run Examples -count=1

verify: build
	./$(BINARY) version
	./$(BINARY) validate -i examples/self-service.kdl --json | grep -q '"ok": true'
	./$(BINARY) render -i examples/self-service.kdl -o dist/smoke --all
	@for f in dist/smoke.*; do test -s "$$f" || (echo "missing $$f" && exit 1); done
	@echo "verify ok ($(VERSION))"

dist:
	@mkdir -p dist
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/sceno-darwin-arm64 $(CMD)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/sceno-darwin-amd64 $(CMD)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/sceno-linux-amd64 $(CMD)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/sceno-linux-arm64 $(CMD)

bump-patch:
	@./scripts/bump-version.sh patch

bump-minor:
	@./scripts/bump-version.sh minor

bump-major:
	@./scripts/bump-version.sh major

release-tag:
	@test -n "$(VERSION)" || (echo "VERSION file missing" && exit 1)
	@git diff --quiet internal/version/VERSION || (echo "Commit VERSION changes first" && exit 1)
	git tag -a "v$(VERSION)" -m "Release v$(VERSION)"
	@echo "Created tag v$(VERSION)"
	@echo "Push with: git push origin main && git push origin v$(VERSION)"

# macOS Apple Silicon release binary (local)
darwin-arm64:
	@mkdir -p dist
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/sceno-darwin-arm64 $(CMD)
