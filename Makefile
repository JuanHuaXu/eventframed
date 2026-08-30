.PHONY: all build test check fmt-check benchmark evaluation-check experiment claims-experiment clean

VERSION ?= 0.1.0-rc.0
GO_BUILD_FLAGS := -trimpath -buildvcs=false
GO_LDFLAGS := -s -w -buildid= -X main.version=$(VERSION)

all: check

build:
	go build $(GO_BUILD_FLAGS) -ldflags "$(GO_LDFLAGS)" -o bin/eventframed ./cmd/eventframed
	cd plugin && npm run build

test:
	go test ./...
	cd plugin && npm test

fmt-check:
	@test -z "$$(gofmt -l $$(find cmd internal benchmark -name '*.go' -type f))" || \
		{ echo "Go files require gofmt:"; gofmt -l $$(find cmd internal benchmark -name '*.go' -type f); exit 1; }

check: fmt-check
	go test -race ./...
	go vet ./...
	cd plugin && npm run check && npm test

benchmark:
	go test ./benchmark -run '^$$' -bench . -benchmem -benchtime=1s -count=5

evaluation-check:
	go test ./internal/evaluation ./cmd/eventframe-eval

experiment:
	go run ./cmd/eventframe-experiment

claims-experiment:
	go run ./cmd/eventframe-claims-experiment

clean:
	rm -rf bin plugin/dist
