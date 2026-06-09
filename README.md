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

## Product Design

The advanced traditional astrology product specification is documented in:

- `docs/product_spec.md`
- `docs/database_schema.sql`

## Swiss Ephemeris

This project uses `github.com/tejzpr/go-swisseph`, which compiles the Swiss
Ephemeris C sources through cgo. It does not need a local desktop `libswe.so`.

Run the terminal smoke test with:

```sh
make sweph-smoke
```

Swiss Ephemeris has separate license terms. Review
`third_party/swisseph/licenses/LICENSE` before deciding how the application will
be distributed.

## Desktop UI

The desktop app uses Fyne:

```sh
make run
```

The first desktop view is a natal workbench:

- `File` menu: new chart window, edit birth data, calculate, save, quit
- `View` menu: open traditional settings and switch between dark and light mode
- `Tools` menu: run the Swiss Ephemeris smoke test
- Toolbar: new chart, edit birth data, settings, calculate, save, refresh, reset, and inspect Swiss Ephemeris
- Left panel: chart library selector
- Center panel: visual chart wheel with in-canvas position readout

Saved charts are stored locally in SQLite at:

```text
~/.config/astro-go/charts.sqlite
```

Use the left `Chart Library` sidebar to switch between stored natal chart inputs,
update the selected saved chart, save the active chart as a new library entry,
or delete a selected chart. New charts created from the sidebar or File menu are
saved automatically after successful calculation. On first startup after the database migration, the
app imports any charts that were previously stored in Fyne preferences under the
app ID `com.esocode.astro-go`.

Birth data is edited in a separate window through `File > Edit Birth Data`, the toolbar
account icon, or the `Edit Birth Data` menu item. Natal chart dialogs now support
place-name lookup through OpenStreetMap Nominatim and can populate latitude and
longitude automatically.

Traditional settings live in a separate window through `View > Settings` or the
toolbar settings icon. House system selection now supports every Swiss Ephemeris
house code exposed by `swe_house_name()`:

- `P` Placidus
- `K` Koch
- `O` Porphyry
- `R` Regiomontanus
- `C` Campanus
- `A` Equal
- `E` Equal (MC)
- `V` Vehlow Equal
- `W` Whole Sign
- `B` Alcabitius
- `T` Topocentric (Polich/Page)
- `M` Morinus
- `U` Krusinski-Pisa-Goelzer
- `H` Horizon / Azimuth
- `X` Meridian
- `Y` APC
- `G` Gauquelin Sectors

On Debian/Ubuntu systems, Fyne may need the Xxf86vm development package:

```sh
sudo apt install libxxf86vm-dev
```

This workstation already had the runtime library but not the development
symlink, so `third_party/system/lib/libXxf86vm.so` points to the installed
runtime library for local builds.

## Fonts

Application fonts live in `internal/assets/fonts` and are embedded into the
binary. `HamburgSymbols.ttf` is used for zodiac and planet glyphs in the chart
wheel. `courier.ttf` is used for compact coordinate labels in the wheel.

The HamburgSymbols reference PDF is kept at `docs/fonts/HamburgSymbols.pdf`.
