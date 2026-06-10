package app

import (
	"math"
	"testing"
	"time"

	"astro-go/internal/astro"
	"astro-go/internal/storage"
)

type mockCalculator struct {
	natalFn func(data astro.BirthData) (astro.Chart, error)
}

func (m mockCalculator) NatalChart(data astro.BirthData) (astro.Chart, error) {
	if m.natalFn != nil {
		return m.natalFn(data)
	}
	return astro.Chart{}, nil
}

func TestResolveCompositeChart(t *testing.T) {
	mock := mockCalculator{
		natalFn: func(data astro.BirthData) (astro.Chart, error) {
			if data.Name == "Base" {
				return astro.Chart{
					Name: "Base",
					Planets: []astro.PlanetPosition{
						{Planet: astro.Sun, Longitude: 10.0},
						{Planet: astro.Moon, Longitude: 350.0},
					},
					Ascendant: astro.Angle{Longitude: 40.0},
					MC:        astro.Angle{Longitude: 130.0},
				}, nil
			}
			return astro.Chart{
				Name: "Compare",
				Planets: []astro.PlanetPosition{
					{Planet: astro.Sun, Longitude: 20.0},
					{Planet: astro.Moon, Longitude: 10.0},
				},
				Ascendant: astro.Angle{Longitude: 60.0},
				MC:        astro.Angle{Longitude: 150.0},
			}, nil
		},
	}

	resolver := NewChartResolver(mock)
	saved := storage.SavedChart{
		ID:                "composite-1",
		Name:              "Composite Chart",
		ChartType:         string(astro.ChartTypeComposite),
		BaseChartID:       "base-id",
		ComparisonChartID: "compare-id",
		LocalDate:         "2020-01-01",
		LocalTime:         "12:00",
		UTCOffset:         "0",
		LatitudeDegrees:   "52.0",
		LongitudeDegrees:  "4.0",
	}

	charts := []storage.SavedChart{
		{
			ID:               "base-id",
			Name:             "Base",
			ChartType:        string(astro.ChartTypeNatal),
			LocalDate:        "2020-01-01",
			LocalTime:        "12:00",
			UTCOffset:        "0",
			LatitudeDegrees:  "52.0",
			LongitudeDegrees: "4.0",
		},
		{
			ID:               "compare-id",
			Name:             "Compare",
			ChartType:        string(astro.ChartTypeNatal),
			LocalDate:        "2020-01-01",
			LocalTime:        "12:00",
			UTCOffset:        "0",
			LatitudeDegrees:  "52.0",
			LongitudeDegrees: "4.0",
		},
	}

	resolved, err := resolver.Resolve(saved, charts)
	if err != nil {
		t.Fatalf("failed to resolve composite chart: %v", err)
	}

	if resolved.Single == nil {
		t.Fatal("expected composite to be a Single chart, got nil")
	}

	composite := resolved.Single
	if composite.ChartType != astro.ChartTypeComposite {
		t.Errorf("expected chart type %s, got %s", astro.ChartTypeComposite, composite.ChartType)
	}

	var compositeSun, compositeMoon float64
	for _, p := range composite.Planets {
		if p.Planet == astro.Sun {
			compositeSun = p.Longitude
		}
		if p.Planet == astro.Moon {
			compositeMoon = p.Longitude
		}
	}

	// Midpoint of 10 and 20 is 15
	if math.Abs(compositeSun-15.0) > 1e-9 {
		t.Errorf("expected Composite Sun at 15.0, got %f", compositeSun)
	}

	// Midpoint of 350 and 10 across 0 degrees is 0.0
	if math.Abs(compositeMoon-0.0) > 1e-9 {
		t.Errorf("expected Composite Moon at 0.0, got %f", compositeMoon)
	}

	// Midpoint of 40 and 60 is 50
	if math.Abs(composite.Ascendant.Longitude-50.0) > 1e-9 {
		t.Errorf("expected Composite Ascendant at 50.0, got %f", composite.Ascendant.Longitude)
	}

	// Midpoint of 130 and 150 is 140
	if math.Abs(composite.MC.Longitude-140.0) > 1e-9 {
		t.Errorf("expected Composite MC at 140.0, got %f", composite.MC.Longitude)
	}
}

func TestResolveDavisonChart(t *testing.T) {
	var capturedData astro.BirthData
	mock := mockCalculator{
		natalFn: func(data astro.BirthData) (astro.Chart, error) {
			capturedData = data
			return astro.Chart{
				Name:      "Davison Result",
				ChartType: astro.ChartTypeDavison,
			}, nil
		},
	}

	resolver := NewChartResolver(mock)
	saved := storage.SavedChart{
		ID:                "davison-1",
		Name:              "Davison Chart",
		ChartType:         string(astro.ChartTypeDavison),
		BaseChartID:       "base-id",
		ComparisonChartID: "compare-id",
		LocalDate:         "2020-01-01",
		LocalTime:         "12:00",
		UTCOffset:         "0",
		LatitudeDegrees:   "52.0",
		LongitudeDegrees:  "4.0",
	}

	charts := []storage.SavedChart{
		{
			ID:               "base-id",
			Name:             "Base",
			ChartType:        string(astro.ChartTypeNatal),
			LocalDate:        "2000-01-01",
			LocalTime:        "12:00",
			UTCOffset:        "0", // UTC
			LatitudeDegrees:  "40.0",
			LongitudeDegrees: "-10.0",
		},
		{
			ID:               "compare-id",
			Name:             "Compare",
			ChartType:        string(astro.ChartTypeNatal),
			LocalDate:        "2000-01-03",
			LocalTime:        "12:00",
			UTCOffset:        "0", // UTC
			LatitudeDegrees:  "50.0",
			LongitudeDegrees: "10.0",
		},
	}

	resolved, err := resolver.Resolve(saved, charts)
	if err != nil {
		t.Fatalf("failed to resolve Davison chart: %v", err)
	}

	if resolved.Single == nil {
		t.Fatal("expected Davison to be a Single chart, got nil")
	}

	// Davison date should be midpoint of 2000-01-01 12:00 and 2000-01-03 12:00, which is 2000-01-02 12:00
	expectedTime := time.Date(2000, 1, 2, 12, 0, 0, 0, time.UTC)
	if !capturedData.DateTimeUTC.Equal(expectedTime) {
		t.Errorf("expected Davison time %s, got %s", expectedTime, capturedData.DateTimeUTC)
	}

	// Midpoint latitude of 40 and 50 is 45
	if math.Abs(capturedData.LatitudeDegrees-45.0) > 1e-9 {
		t.Errorf("expected Davison latitude 45.0, got %f", capturedData.LatitudeDegrees)
	}

	// Midpoint longitude of -10 and 10 is 0
	if math.Abs(capturedData.LongitudeDegrees-0.0) > 1e-9 {
		t.Errorf("expected Davison longitude 0.0, got %f", capturedData.LongitudeDegrees)
	}
}

func TestResolvePrimaryDirections(t *testing.T) {
	mock := mockCalculator{
		natalFn: func(data astro.BirthData) (astro.Chart, error) {
			return astro.Chart{
				Planets: []astro.PlanetPosition{
					{Planet: astro.Sun, Longitude: 100.0},
				},
				Ascendant: astro.Angle{Longitude: 45.0},
				MC:        astro.Angle{Longitude: 135.0},
			}, nil
		},
	}

	resolver := NewChartResolver(mock)
	saved := storage.SavedChart{
		ID:           "pd-1",
		Name:         "Primary Directions",
		ChartType:    string(astro.ChartTypePrimaryDirection),
		BaseChartID:  "base-id",
		ReferenceUTC: "2010-01-01T12:00:00Z",
		DirectionKey: "naibod",
	}

	charts := []storage.SavedChart{
		{
			ID:               "base-id",
			Name:             "Base",
			ChartType:        string(astro.ChartTypeNatal),
			LocalDate:        "2000-01-01",
			LocalTime:        "12:00",
			UTCOffset:        "0",
			LatitudeDegrees:  "52.0",
			LongitudeDegrees: "4.0",
		},
	}

	resolved, err := resolver.Resolve(saved, charts)
	if err != nil {
		t.Fatalf("failed to resolve Primary Directions: %v", err)
	}

	if resolved.Single == nil {
		t.Fatal("expected Primary Directions to be resolved as Single, got nil")
	}

	single := resolved.Single
	var pdSun float64
	for _, p := range single.Planets {
		if p.Planet == astro.Sun {
			pdSun = p.Longitude
		}
	}

	// Years = 10 (plus leap days: 2000, 2004, 2008). 3653 days / 365.242199 = 10.0015825 years.
	duration := time.Date(2010, 1, 1, 12, 0, 0, 0, time.UTC).Sub(time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC))
	years := float64(duration) / (365.242199 * 24 * float64(time.Hour))
	expectedSun := astro.NormalizeDegrees(100.0 + (years * 0.9855556))
	if math.Abs(pdSun-expectedSun) > 1e-9 {
		t.Errorf("expected Directed Sun at %f, got %f", expectedSun, pdSun)
	}
}

func TestResolveRelocatedReturn(t *testing.T) {
	var capturedData astro.BirthData
	mock := mockCalculator{
		natalFn: func(data astro.BirthData) (astro.Chart, error) {
			capturedData = data
			return astro.Chart{
				Planets: []astro.PlanetPosition{
					{Planet: astro.Sun, Longitude: 120.0},
				},
			}, nil
		},
	}

	resolver := NewChartResolver(mock)
	saved := storage.SavedChart{
		ID:                    "return-1",
		Name:                  "Solar Return",
		ChartType:             string(astro.ChartTypeSolarReturn),
		BaseChartID:           "base-id",
		ReferenceUTC:          "2026-01-01T12:00:00Z",
		RelocatedLatitude:     "40.7128",
		RelocatedLongitude:    "-74.0060",
		RelocatedLocationName: "New York, USA",
	}

	charts := []storage.SavedChart{
		{
			ID:               "base-id",
			Name:             "Base",
			ChartType:        string(astro.ChartTypeNatal),
			LocalDate:        "2000-01-01",
			LocalTime:        "12:00",
			UTCOffset:        "0",
			LatitudeDegrees:  "52.3676",
			LongitudeDegrees: "4.9041",
			LocationName:     "Amsterdam, Netherlands",
		},
	}

	_, _ = resolver.Resolve(saved, charts)

	// In the calculation phase, the calculator should be called with the relocated coordinates
	if math.Abs(capturedData.LatitudeDegrees-40.7128) > 1e-4 {
		t.Errorf("expected calculation latitude 40.7128, got %f", capturedData.LatitudeDegrees)
	}
	if math.Abs(capturedData.LongitudeDegrees - -74.0060) > 1e-4 {
		t.Errorf("expected calculation longitude -74.0060, got %f", capturedData.LongitudeDegrees)
	}
	if capturedData.LocationName != "New York, USA" {
		t.Errorf("expected LocationName 'New York, USA', got %q", capturedData.LocationName)
	}
}

func TestResolveSynastryWithSolarReturn(t *testing.T) {
	mock := mockCalculator{
		natalFn: func(data astro.BirthData) (astro.Chart, error) {
			if data.Name == "Base Natal" {
				return astro.Chart{
					Name: "Base Natal",
					Planets: []astro.PlanetPosition{
						{Planet: astro.Sun, Longitude: 100.0},
					},
				}, nil
			}
			return astro.Chart{
				Name: "Solar Return 2026",
				Planets: []astro.PlanetPosition{
					{Planet: astro.Sun, Longitude: 100.0},
					{Planet: astro.Moon, Longitude: 240.0},
				},
			}, nil
		},
	}

	resolver := NewChartResolver(mock)

	charts := []storage.SavedChart{
		{
			ID:               "natal-id",
			Name:             "Base Natal",
			ChartType:        string(astro.ChartTypeNatal),
			LocalDate:        "2000-01-01",
			LocalTime:        "12:00",
			UTCOffset:        "0",
			LatitudeDegrees:  "52.0",
			LongitudeDegrees: "4.0",
		},
		{
			ID:           "return-id",
			Name:         "Solar Return 2026",
			ChartType:    string(astro.ChartTypeSolarReturn),
			BaseChartID:  "natal-id",
			ReferenceUTC: "2026-01-01T12:00:00Z",
		},
	}

	synastrySaved := storage.SavedChart{
		ID:                "synastry-id",
		Name:              "Natal + Return Comparison",
		ChartType:         string(astro.ChartTypeSynastry),
		BaseChartID:       "natal-id",
		ComparisonChartID: "return-id",
	}

	resolved, err := resolver.Resolve(synastrySaved, charts)
	if err != nil {
		t.Fatalf("failed to resolve Synastry with Solar Return: %v", err)
	}

	if resolved.Synastry == nil {
		t.Fatal("expected Synastry chart to be resolved, got nil")
	}

	synastry := resolved.Synastry
	if synastry.InnerChart.Name != "Base Natal" {
		t.Errorf("expected inner chart to be 'Base Natal', got %q", synastry.InnerChart.Name)
	}
	if synastry.OuterChart.Name != "Solar Return 2026" {
		t.Errorf("expected outer chart to be 'Solar Return 2026', got %q", synastry.OuterChart.Name)
	}

	var outerMoon float64
	for _, p := range synastry.OuterChart.Planets {
		if p.Planet == astro.Moon {
			outerMoon = p.Longitude
		}
	}
	if math.Abs(outerMoon-240.0) > 1e-9 {
		t.Errorf("expected outer chart moon position to be 240.0, got %f", outerMoon)
	}
}
