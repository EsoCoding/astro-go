package timezone

import (
	"testing"
)

func TestLookupTimezone(t *testing.T) {
	tests := []struct {
		name string
		lat  float64
		lon  float64
		want string
	}{
		{"Amsterdam", 52.3676, 4.9041, "Europe/Amsterdam"},
		{"New York", 40.7128, -74.0060, "America/New_York"},
		{"Tokyo", 35.6762, 139.6503, "Asia/Tokyo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LookupTimezone(tt.lat, tt.lon)
			if got != tt.want {
				t.Errorf("LookupTimezone() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCalculateOffset(t *testing.T) {
	tests := []struct {
		name    string
		tzName  string
		dateStr string
		timeStr string
		want    float64
		wantErr bool
	}{
		{"Amsterdam Summer (DST)", "Europe/Amsterdam", "2026-06-08", "12:00", 2.0, false},
		{"Amsterdam Winter (ST)", "Europe/Amsterdam", "2026-12-08", "12:00", 1.0, false},
		{"New York Summer (DST)", "America/New_York", "2026-06-08", "12:00", -4.0, false},
		{"New York Winter (ST)", "America/New_York", "2026-12-08", "12:00", -5.0, false},
		{"Tokyo Summer (ST, no DST)", "Asia/Tokyo", "2026-06-08", "12:00", 9.0, false},
		{"Tokyo Winter (ST, no DST)", "Asia/Tokyo", "2026-12-08", "12:00", 9.0, false},
		{"Invalid Timezone", "Invalid/Timezone", "2026-06-08", "12:00", 0, true},
		{"Invalid Date Format", "Europe/Amsterdam", "2026/06/08", "12:00", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateOffset(tt.tzName, tt.dateStr, tt.timeStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateOffset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("CalculateOffset() = %g, want %g", got, tt.want)
			}
		})
	}
}
