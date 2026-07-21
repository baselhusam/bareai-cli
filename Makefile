.PHONY: all build test test-integration smoke smoke-windows lint run clean release-snapshot goreleaser-check

BINARY := bareai
CMD := ./cmd/bareai

all: build test

build:
	go build -o $(BINARY) $(CMD)

test:
	go test ./...

test-integration: build
	go test -tags=integration ./internal/smoke/...

smoke: build
	./scripts/smoke.sh

smoke-windows: build
	powershell -ExecutionPolicy Bypass -File scripts/smoke.ps1

lint:
	golangci-lint run

run:
	go run $(CMD) -- $(ARGS)

release-snapshot:
	goreleaser release --snapshot --clean

goreleaser-check:
	goreleaser check

clean:
	rm -f $(BINARY) $(BINARY).exe
	rm -rf dist/
