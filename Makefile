APP := astro-go
SWEPH_LIB := $(CURDIR)/third_party/swisseph/lib
SYSTEM_LIB := $(CURDIR)/third_party/system/lib
SWEPH_CGO_LDFLAGS := -L$(SWEPH_LIB) -Wl,-rpath,$(SWEPH_LIB) -L$(SYSTEM_LIB)

.PHONY: build run test fmt env sweph-smoke desktop

build:
	CGO_LDFLAGS="$(SWEPH_CGO_LDFLAGS)" go build -o bin/$(APP) ./cmd/$(APP)

run:
	CGO_LDFLAGS="$(SWEPH_CGO_LDFLAGS)" go run ./cmd/$(APP)

desktop: run

test:
	CGO_LDFLAGS="$(SWEPH_CGO_LDFLAGS)" go test ./...

sweph-smoke:
	CGO_LDFLAGS="$(SWEPH_CGO_LDFLAGS)" go run ./cmd/sweph-smoke

fmt:
	go fmt ./...

env:
	go env GOPATH GOROOT GOENV GOMODCACHE
