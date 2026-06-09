package sweph

import (
	"fmt"

	swisseph "github.com/tejzpr/go-swisseph"
)

type SmokeResult struct {
	Version      string
	JulianDay    float64
	SunLongitude float64
	SunLatitude  float64
}

func Smoke() (SmokeResult, error) {
	version := swisseph.Version()
	defer swisseph.Close()

	configureEphemerisPath()

	julianDay := swisseph.Julday(2026, 6, 7, 12, swisseph.GregCal)
	result := swisseph.CalcUT(julianDay, swisseph.Sun, swisseph.FlagSwieph|swisseph.FlagSpeed)
	if result.Flag < 0 {
		return SmokeResult{}, fmt.Errorf("swiss ephemeris calculation failed: %s", result.Error)
	}

	return SmokeResult{
		Version:      version,
		JulianDay:    julianDay,
		SunLongitude: result.Data[0],
		SunLatitude:  result.Data[1],
	}, nil
}
