package sweph

import (
	"testing"
	"time"

	"astro-go/internal/astro"
)

func TestNatalChartCalculatesTraditionalPlanets(t *testing.T) {
	chart, err := NewCalculator().NatalChart(astro.BirthData{
		Name:             "Test Chart",
		DateTimeUTC:      time.Date(1990, 1, 1, 12, 0, 0, 0, time.UTC),
		LatitudeDegrees:  52.3676,
		LongitudeDegrees: 4.9041,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(chart.Planets) < len(astro.TraditionalPlanets) {
		t.Errorf("got %d planets, want at least %d", len(chart.Planets), len(astro.TraditionalPlanets))
	}
	if len(chart.Houses) != 12 {
		t.Fatalf("got %d houses, want 12", len(chart.Houses))
	}
	if chart.Ascendant.Longitude < 0 || chart.Ascendant.Longitude >= 360 {
		t.Fatalf("ascendant longitude out of range: %f", chart.Ascendant.Longitude)
	}
}
