package app

import (
	"testing"
	"time"

	"astro-go/internal/astro"
	"astro-go/internal/storage"
)

type fakeChartCalculator struct {
	lastBirthData astro.BirthData
	result        astro.Chart
	err           error
}

func (f *fakeChartCalculator) NatalChart(data astro.BirthData) (astro.Chart, error) {
	f.lastBirthData = data
	if f.err != nil {
		return astro.Chart{}, f.err
	}
	result := f.result
	result.Name = data.Name
	result.DateTimeUTC = data.DateTimeUTC
	result.Latitude = data.LatitudeDegrees
	result.Longitude = data.LongitudeDegrees
	return result, nil
}

func TestChartResolverResolvesNatalSavedChart(t *testing.T) {
	calc := &fakeChartCalculator{result: astro.Chart{}}
	resolver := NewChartResolver(calc)

	saved := storage.SavedChart{
		Name:             "Natal Example",
		ChartType:        string(astro.ChartTypeNatal),
		LocalDate:        "1990-01-01",
		LocalTime:        "12:00",
		UTCOffset:        "0",
		LatitudeDegrees:  "52.3676",
		LongitudeDegrees: "4.9041",
	}

	resolved, err := resolver.Resolve(saved, nil)
	if err != nil {
		t.Fatalf("Resolve(natal) error = %v", err)
	}
	if resolved.Single == nil {
		t.Fatal("resolved.Single = nil for natal chart")
	}
	chart := *resolved.Single
	if chart.Name != "Natal Example" {
		t.Fatalf("chart.Name = %q, want natal name", chart.Name)
	}
	if calc.lastBirthData.DateTimeUTC != time.Date(1990, 1, 1, 12, 0, 0, 0, time.UTC) {
		t.Fatalf("natal UTC = %s, want 1990-01-01 12:00 UTC", calc.lastBirthData.DateTimeUTC)
	}
}

func TestChartResolverResolvesTransitSavedChart(t *testing.T) {
	calc := &fakeChartCalculator{result: astro.Chart{}}
	resolver := NewChartResolver(calc)

	base := storage.SavedChart{
		ID:               "base",
		Name:             "Alice Natal",
		ChartType:        string(astro.ChartTypeNatal),
		LocalDate:        "1990-01-01",
		LocalTime:        "12:00",
		UTCOffset:        "0",
		LatitudeDegrees:  "52.3676",
		LongitudeDegrees: "4.9041",
	}
	transit := storage.SavedChart{
		Name:          "Alice Transit",
		ChartType:     string(astro.ChartTypeTransit),
		BaseChartID:   "base",
		ReferenceDate: "2026-06-08",
		ReferenceTime: "09:30",
	}

	resolved, err := resolver.Resolve(transit, []storage.SavedChart{base})
	if err != nil {
		t.Fatalf("Resolve(transit) error = %v", err)
	}
	if resolved.Synastry == nil {
		t.Fatal("resolved.Synastry = nil for transit chart")
	}
	synastry := *resolved.Synastry
	if synastry.Name != "Alice Transit" {
		t.Fatalf("synastry.Name = %q, want transit definition name", synastry.Name)
	}
	if synastry.InnerChart.Name != "Alice Natal" {
		t.Fatalf("synastry.InnerChart.Name = %q, want Alice Natal", synastry.InnerChart.Name)
	}
	if synastry.OuterChart.Name != "Transits" {
		t.Fatalf("synastry.OuterChart.Name = %q, want Transits", synastry.OuterChart.Name)
	}
	if calc.lastBirthData.DateTimeUTC != time.Date(2026, 6, 8, 9, 30, 0, 0, time.UTC) {
		t.Fatalf("transit UTC = %s, want 2026-06-08 09:30 UTC", calc.lastBirthData.DateTimeUTC)
	}
}

func TestChartResolverResolvesProgressionSavedChart(t *testing.T) {
	calc := &fakeChartCalculator{result: astro.Chart{}}
	resolver := NewChartResolver(calc)

	base := storage.SavedChart{
		ID:               "base",
		Name:             "Alice Natal",
		ChartType:        string(astro.ChartTypeNatal),
		LocalDate:        "1990-01-01",
		LocalTime:        "12:00",
		UTCOffset:        "0",
		LatitudeDegrees:  "52.3676",
		LongitudeDegrees: "4.9041",
	}
	prog := storage.SavedChart{
		Name:          "Alice Progressed",
		ChartType:     string(astro.ChartTypeSecondaryProgression),
		BaseChartID:   "base",
		ReferenceDate: "2026-06-08",
		ReferenceTime: "09:30",
	}

	resolved, err := resolver.Resolve(prog, []storage.SavedChart{base})
	if err != nil {
		t.Fatalf("Resolve(progression) error = %v", err)
	}
	if resolved.Synastry == nil {
		t.Fatal("resolved.Synastry = nil for progression chart")
	}
	synastry := *resolved.Synastry
	if synastry.Name != "Alice Progressed" {
		t.Fatalf("synastry.Name = %q, want progression definition name", synastry.Name)
	}
}

func TestChartResolverResolvesSynastrySavedChart(t *testing.T) {
	calc := &fakeChartCalculator{result: astro.Chart{}}
	resolver := NewChartResolver(calc)

	inner := storage.SavedChart{
		ID:               "inner",
		Name:             "Alice",
		ChartType:        string(astro.ChartTypeNatal),
		LocalDate:        "1990-01-01",
		LocalTime:        "12:00",
		UTCOffset:        "0",
		LatitudeDegrees:  "52.3676",
		LongitudeDegrees: "4.9041",
	}
	outer := storage.SavedChart{
		ID:               "outer",
		Name:             "Bob",
		ChartType:        string(astro.ChartTypeNatal),
		LocalDate:        "1991-02-03",
		LocalTime:        "14:30",
		UTCOffset:        "1",
		LatitudeDegrees:  "51.9244",
		LongitudeDegrees: "4.4777",
	}
	synastry := storage.SavedChart{
		Name:              "Alice x Bob",
		ChartType:         string(astro.ChartTypeSynastry),
		BaseChartID:       "inner",
		ComparisonChartID: "outer",
	}

	resolved, err := resolver.Resolve(synastry, []storage.SavedChart{inner, outer})
	if err != nil {
		t.Fatalf("Resolve(synastry) error = %v", err)
	}
	if resolved.Synastry == nil {
		t.Fatal("resolved.Synastry = nil for synastry chart")
	}
	if resolved.Synastry.InnerChart.Name != "Alice" || resolved.Synastry.OuterChart.Name != "Bob" {
		t.Fatalf("synastry names = %q / %q, want Alice / Bob", resolved.Synastry.InnerChart.Name, resolved.Synastry.OuterChart.Name)
	}
}

func TestChartResolverResolvesSolarReturnSavedChart(t *testing.T) {
	calc := &fakeChartCalculator{result: astro.Chart{
		Planets: []astro.PlanetPosition{
			{Planet: astro.Sun, Longitude: 100.0},
			{Planet: astro.Moon, Longitude: 200.0},
		},
	}}
	resolver := NewChartResolver(calc)

	base := storage.SavedChart{
		ID:               "base",
		Name:             "Alice Natal",
		ChartType:        string(astro.ChartTypeNatal),
		LocalDate:        "1990-01-01",
		LocalTime:        "12:00",
		UTCOffset:        "0",
		LatitudeDegrees:  "52.3676",
		LongitudeDegrees: "4.9041",
	}
	sr := storage.SavedChart{
		Name:          "Alice Solar Return",
		ChartType:     string(astro.ChartTypeSolarReturn),
		BaseChartID:   "base",
		ReferenceDate: "2026-06-08",
		ReferenceTime: "09:30",
	}

	resolved, err := resolver.Resolve(sr, []storage.SavedChart{base})
	if err != nil {
		t.Fatalf("Resolve(solar_return) error = %v", err)
	}
	if resolved.Synastry == nil {
		t.Fatal("resolved.Synastry = nil for solar return chart")
	}
}

func TestChartResolverResolvesLunarReturnSavedChart(t *testing.T) {
	calc := &fakeChartCalculator{result: astro.Chart{
		Planets: []astro.PlanetPosition{
			{Planet: astro.Sun, Longitude: 100.0},
			{Planet: astro.Moon, Longitude: 200.0},
		},
	}}
	resolver := NewChartResolver(calc)

	base := storage.SavedChart{
		ID:               "base",
		Name:             "Alice Natal",
		ChartType:        string(astro.ChartTypeNatal),
		LocalDate:        "1990-01-01",
		LocalTime:        "12:00",
		UTCOffset:        "0",
		LatitudeDegrees:  "52.3676",
		LongitudeDegrees: "4.9041",
	}
	lr := storage.SavedChart{
		Name:          "Alice Lunar Return",
		ChartType:     string(astro.ChartTypeLunarReturn),
		BaseChartID:   "base",
		ReferenceDate: "2026-06-08",
		ReferenceTime: "09:30",
	}

	resolved, err := resolver.Resolve(lr, []storage.SavedChart{base})
	if err != nil {
		t.Fatalf("Resolve(lunar_return) error = %v", err)
	}
	if resolved.Synastry == nil {
		t.Fatal("resolved.Synastry = nil for lunar return chart")
	}
}
