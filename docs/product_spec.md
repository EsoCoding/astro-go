# Astro Go Product Specification

## Product Goal

Astro Go is a desktop-first traditional astrology workbench for serious chart
calculation, judgment, timing, research, and reporting. The product prioritizes
traditional doctrine and transparent rule-based reasoning. Modern planets and
modern interpretive tools are optional modules, not the default foundation.

## Assumptions

- The first production target is a Linux desktop application built with Go and Fyne.
- Swiss Ephemeris remains the primary astronomical engine.
- SQLite is the default local database.
- Web deployment is future architecture, not part of the first production release.
- Traditional astrology features ship first; modern features are plugin/configuration modules.

## Feature List

### Chart Modules

- Natal charts
- Horary charts
- Electional charts
- Mundane and ingress charts
- Solar and lunar returns
- Annual profections
- Transits over natal charts
- Synastry and relationship comparison
- Secondary progressions
- Zodiacal releasing
- Primary directions as a later advanced module

### Astronomical Calculations

- Planet longitude, latitude, speed, retrogradation, declination, right ascension
- Ascendant, MC, house cusps, lots, fixed star contacts
- Tropical and sidereal zodiacs
- Whole Sign, Regiomontanus, Placidus, Alcabitius, Porphyry, Campanus, Equal, Meridian
- Time zone and daylight saving handling
- Manual coordinates first, geocoding later
- Julian/Gregorian historical calendar support as a dedicated module

### Traditional Rules

- Essential dignities: domicile, exaltation, triplicity, term, face
- Accidental dignities: angularity, speed, sect, joy, visibility, combustion, cazimi
- Debilities: detriment, fall, peregrine, cadent, retrograde, combust, besieged
- Reception: mutual, mixed, and dignity-specific reception
- Aspects: Ptolemaic, applying/separating, dexter/sinister, moiety, whole-sign, partile
- Sect analysis
- Lots: Fortune, Spirit, Eros, Necessity, Victory, Nemesis, Courage, Basis, custom lots
- Fixed stars, lunar mansions, planetary hours/days
- Void of course Moon
- Translation and collection of light
- Prohibition, frustration, refranation, besiegement
- Almuten, Hyleg, Alcocoden
- Egyptian, Ptolemaic, and Chaldean bounds
- Antiscia, contra-antiscia, dodecatemoria, monomoiria

## GUI Wireframe

### Desktop Layout

- Left sidebar: chart type selector, project folders, saved chart library, settings shortcut
- Top toolbar: new, open, save, save as, export, calculate, compare, settings
- Center panel: interactive chart wheel
- Right panel: tabbed analysis for planets, houses, aspects, dignities, condition, interpretation
- Bottom timeline: transits, progressions, directions, lunar aspects, event timing

### Interaction Requirements

- Chart wheel zoom and pan
- Hover tooltips for planets, signs, houses, lots, fixed stars, aspects
- Planet selection drives the right analysis panel
- Toggles for modern planets, aspect lines, house cusps, lots, fixed stars, antiscia
- Dark and light themes
- Keyboard shortcuts for calculate, save, new chart, export, search
- Export PDF, PNG, SVG, JSON, CSV

## Recommended Tech Stack

- Language: Go
- GUI: Fyne
- Database: SQLite via `modernc.org/sqlite`
- Ephemeris: Swiss Ephemeris via `github.com/mshafiee/swephgo`
- Chart rendering: Fyne custom canvas widgets, later shared renderer abstraction for PDF/SVG
- PDF export: Go PDF generator such as `gofpdf` or `unidoc` after report layout is stable
- Testing: Go unit tests, regression fixture tests, screenshot tests for chart rendering
- Packaging: Makefile for local builds, Fyne packaging for desktop installers

## Modular Architecture

```text
cmd/
  astro-go/          desktop application entrypoint
  sweph-smoke/       terminal Swiss Ephemeris smoke test
internal/
  astro/             domain models and traditional rules
  sweph/             Swiss Ephemeris adapter
  storage/           SQLite repositories and migrations
  ui/                Fyne application shell and widgets
  report/            report/export generation
  interpretation/    transparent reasoning engine
  plugins/           optional rule/render/export modules
docs/
  product_spec.md
  database_schema.sql
```

## Calculation Engine

The calculation engine should stay separate from UI and storage. It accepts a
typed chart request and returns a typed calculated chart.

Core request fields:

- Chart type
- Local date/time
- Time zone
- Calendar mode
- Latitude/longitude
- Zodiac mode
- House system
- Planet set
- Required modules: lots, fixed stars, mansions, directions, transits

Core result fields:

- Normalized UTC timestamp
- Planet positions and motion data
- House cusps and angles
- Lots and calculated points
- Aspect graph
- Dignity and condition tables
- Calculation metadata and ephemeris version

## Rules Engine

Astrological rules must be explicit data or named functions, never hidden in UI
code. Every rule evaluation should return:

- Rule ID
- Rule label
- Inputs
- Result
- Confidence
- Explanation
- Source tradition or configuration set

Example rule families:

- `essential.domicile`
- `essential.exaltation`
- `essential.bounds.egyptian`
- `accidental.combustion`
- `aspect.applying`
- `horary.perfection.translation`
- `sect.day_night`

## Interpretation Engine

Interpretation is rule-based and evidence-driven. It should not produce vague
horoscope text. Every paragraph in a judgment should be traceable to one or more
rules.

Judgment output structure:

- Summary
- Relevant significators
- Planetary condition
- Aspect/perfection analysis
- Reception analysis
- Timing indicators
- Warnings and considerations
- Alternative readings
- Confidence level
- Technical appendix

## Report Generation

Reports should support:

- Natal analysis
- Horary judgment
- Electional recommendation
- Transit forecast
- Annual profection
- Solar return
- Synastry comparison

Each report includes:

- Chart image
- Planet, house, aspect, dignity, lots, and fixed star tables
- Interpretive summary
- Technical appendix with exact calculation settings

## Implementation Roadmap

### Phase 1: Current Prototype Stabilization

- SQLite chart library
- Responsive chart wheel
- Natal calculation with seven traditional planets
- Whole Sign houses
- Basic dignity and aspect tables
- Separate birth data/settings windows

### Phase 2: Traditional Natal Core

- Full essential dignity tables
- Sect analysis
- Egyptian/Ptolemaic bounds
- Lots of Fortune and Spirit
- Planetary condition scoring with explanations
- Export JSON and CSV

### Phase 3: Horary Module

- Horary chart type
- Considerations before judgment
- Significator assignment
- Moon next aspects
- Perfection, prohibition, frustration, refranation
- Structured horary judgment report

### Phase 4: Timing and Comparison

- Transits over natal
- Solar returns
- Annual profections
- Synastry comparison
- Bottom timeline panel

### Phase 5: Reports and Professional Workflow

- PDF reports
- Projects, clients, tags, notes
- Search
- Chart version history
- Print-ready templates

### Phase 6: Advanced Modules

- Fixed stars
- Lunar mansions
- Zodiacal releasing
- Primary directions
- Plugin API
- Optional cloud sync architecture

## Initial Prototype Code

The current prototype already includes:

- Go/Fyne app entrypoint
- Swiss Ephemeris adapter
- SQLite saved chart store
- Natal chart calculation
- Traditional chart wheel
- Saved chart sidebar
- Basic planets, houses, aspects, and dignity analysis

## Test Examples

- Known natal chart fixtures for planet longitudes
- Dignity table fixtures by sign and degree
- Aspect application/separation fixtures
- Horary examples with known perfection outcomes
- Storage tests for save, update, delete, migration, and schema creation

## Future Roadmap

- Data import/export
- Geocoding
- Historical time zones and calendars
- Plugin marketplace or local plugin folder
- Web UI adapter
- Encrypted client data store
- Collaborative/cloud sync mode
