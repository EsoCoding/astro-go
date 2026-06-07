# astro-go

Astrology application built with Go.

## Go setup

This project currently uses Go `1.26.4` through `gvm`.

```sh
source "$HOME/.gvm/scripts/gvm"
gvm use go1.26.4 --default
go version
```

If a new shell cannot find `go`, make sure this line exists in `~/.bashrc` or `~/.zshrc`:

```sh
[[ -s "$HOME/.gvm/scripts/gvm" ]] && source "$HOME/.gvm/scripts/gvm"
```

## Development

```sh
make run
make test
make build
```

The compiled binary is written to `bin/astro-go`.

## Swiss Ephemeris

This project uses `github.com/mshafiee/swephgo`, which is a cgo binding to the
native Swiss Ephemeris library. The local shared library is stored at:

```text
third_party/swisseph/lib/libswe.so
```

Run the terminal smoke test with:

```sh
make sweph-smoke
```

To make direct `go test ./...` and `go run ./cmd/sweph-smoke` commands work
without the Makefile, configure Go's cgo linker flags once:

```sh
go env -w CGO_LDFLAGS="-L$(pwd)/third_party/swisseph/lib -Wl,-rpath,$(pwd)/third_party/swisseph/lib"
```

Swiss Ephemeris has separate license terms. Review
`third_party/swisseph/licenses/LICENSE` before deciding how the application will
be distributed.

## Desktop UI

The desktop app uses Fyne:

```sh
make run
```

On Debian/Ubuntu systems, Fyne may need the Xxf86vm development package:

```sh
sudo apt install libxxf86vm-dev
```

This workstation already had the runtime library but not the development
symlink, so `third_party/system/lib/libXxf86vm.so` points to the installed
runtime library for local builds.
