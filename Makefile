APP := astro-go
SYSTEM_LIB := $(CURDIR)/third_party/system/lib
CGO_SYSTEM_LDFLAGS := -L$(SYSTEM_LIB)

.PHONY: build run test fmt env sweph-smoke desktop

build:
	CGO_LDFLAGS="$(CGO_SYSTEM_LDFLAGS)" go build -o bin/$(APP) ./cmd/$(APP)

run:
	CGO_LDFLAGS="$(CGO_SYSTEM_LDFLAGS)" go run ./cmd/$(APP)

desktop: run

test:
	CGO_LDFLAGS="$(CGO_SYSTEM_LDFLAGS)" go test ./...

sweph-smoke:
	CGO_LDFLAGS="$(CGO_SYSTEM_LDFLAGS)" go run ./cmd/sweph-smoke

fmt:
	go fmt ./...

env:
	go env GOPATH GOROOT GOENV GOMODCACHE
