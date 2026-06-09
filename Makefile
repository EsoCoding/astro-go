APP := astro-go

.PHONY: build run test fmt env sweph-smoke desktop

build:
	go build -o bin/$(APP) ./cmd/$(APP)

run:
	go run ./cmd/$(APP)

desktop: run

test:
	go test ./...

sweph-smoke:
	go run ./cmd/sweph-smoke

fmt:
	go fmt ./...

env:
	go env GOPATH GOROOT GOENV GOMODCACHE
