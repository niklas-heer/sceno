.PHONY: build test install clean lint examples

BINARY := sceno
CMD := ./cmd/sceno
VERSION ?= dev

build:
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY) $(CMD)

test:
	go test -race -count=1 ./...

install:
	go install -ldflags="-s -w -X main.version=$(VERSION)" $(CMD)

clean:
	rm -f $(BINARY)
	rm -rf dist/

lint:
	go vet ./...
	go test ./...

examples:
	go test ./internal/pipeline/ -run Examples -count=1

# macOS Apple Silicon release binary (local)
darwin-arm64:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(VERSION)" -o dist/sceno-darwin-arm64 $(CMD)
