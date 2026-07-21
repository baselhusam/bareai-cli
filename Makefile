.PHONY: all build test lint run clean

BINARY := bareai
CMD := ./cmd/bareai

all: build test

build:
	go build -o $(BINARY) $(CMD)

test:
	go test ./...

lint:
	golangci-lint run

run:
	go run $(CMD) -- $(ARGS)

clean:
	rm -f $(BINARY) $(BINARY).exe
