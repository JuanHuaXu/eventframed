.PHONY: all build test check clean

all: check

build:
	go build -o bin/eventframed ./cmd/eventframed
	cd plugin && npm run build

test:
	go test ./...
	cd plugin && npm test

check:
	go test -race ./...
	go vet ./...
	cd plugin && npm run check && npm test

clean:
	rm -rf bin plugin/dist
