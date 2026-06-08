package timezone

import (
	"fmt"
	"time"
	_ "time/tzdata"

	"github.com/ringsaturn/tzf"
)

var finder tzf.F

func init() {
	var err error
	finder, err = tzf.NewDefaultFinder()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize timezone finder: %v", err))
	}
}

// LookupTimezone returns the timezone name (e.g. "Europe/Amsterdam") for the given coordinates.
func LookupTimezone(lat, lon float64) string {
	return finder.GetTimezoneName(lon, lat) // Note: tzf uses lon, lat order
}

// CalculateOffset returns the UTC offset in hours (e.g. 1.0 or -5.5) for a local date and time in the given timezone.
func CalculateOffset(timezoneName, dateStr, timeStr string) (float64, error) {
	if timezoneName == "" {
		return 0, fmt.Errorf("empty timezone name")
	}
	loc, err := time.LoadLocation(timezoneName)
	if err != nil {
		return 0, fmt.Errorf("failed to load location %q: %w", timezoneName, err)
	}

	// Parse the local date and time in that location
	localTime, err := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+timeStr, loc)
	if err != nil {
		return 0, fmt.Errorf("failed to parse date/time %q %q: %w", dateStr, timeStr, err)
	}

	_, offsetSeconds := localTime.Zone()
	return float64(offsetSeconds) / 3600.0, nil
}
