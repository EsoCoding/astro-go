package ui

import (
	"testing"
	"time"

	"astro-go/internal/astro"
)

func TestStepTime(t *testing.T) {
	start := time.Date(2024, 1, 31, 12, 30, 15, 0, time.UTC)

	tests := []struct {
		name   string
		unit   string
		amount int
		want   time.Time
	}{
		{
			name:   "second",
			unit:   timeStepUnitSecond,
			amount: 1,
			want:   time.Date(2024, 1, 31, 12, 30, 16, 0, time.UTC),
		},
		{
			name:   "minute",
			unit:   timeStepUnitMinute,
			amount: -30,
			want:   time.Date(2024, 1, 31, 12, 0, 15, 0, time.UTC),
		},
		{
			name:   "hour",
			unit:   timeStepUnitHour,
			amount: 6,
			want:   time.Date(2024, 1, 31, 18, 30, 15, 0, time.UTC),
		},
		{
			name:   "week",
			unit:   timeStepUnitWeek,
			amount: 2,
			want:   time.Date(2024, 2, 14, 12, 30, 15, 0, time.UTC),
		},
		{
			name:   "month uses calendar arithmetic",
			unit:   timeStepUnitMonth,
			amount: 1,
			want:   time.Date(2024, 3, 2, 12, 30, 15, 0, time.UTC),
		},
		{
			name:   "year",
			unit:   timeStepUnitYear,
			amount: 1,
			want:   time.Date(2025, 1, 31, 12, 30, 15, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stepTime(start, tt.unit, tt.amount)
			if !got.Equal(tt.want) {
				t.Fatalf("stepTime() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestSwapSynastry(t *testing.T) {
	inner := astro.Chart{
		Name: "Inner",
		Planets: []astro.PlanetPosition{
			{Planet: astro.Sun, Longitude: 0},
		},
	}
	outer := astro.Chart{
		Name: "Outer",
		Planets: []astro.PlanetPosition{
			{Planet: astro.Moon, Longitude: 180},
		},
	}

	swapped := swapSynastry(astro.SynastryChart{
		Name:       "Pair",
		InnerChart: inner,
		OuterChart: outer,
	})

	if swapped.InnerChart.Name != "Outer" || swapped.OuterChart.Name != "Inner" {
		t.Fatalf("swapSynastry() did not swap charts: %s x %s", swapped.InnerChart.Name, swapped.OuterChart.Name)
	}
	if len(swapped.InterAspects) != 1 {
		t.Fatalf("swapSynastry() aspects = %d, want 1", len(swapped.InterAspects))
	}
	if swapped.InterAspects[0].Inner != astro.Moon || swapped.InterAspects[0].Outer != astro.Sun {
		t.Fatalf("swapSynastry() aspects use wrong order: %#v", swapped.InterAspects[0])
	}
}
