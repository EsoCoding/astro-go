package sweph

import (
	"bytes"
	"fmt"

	"github.com/mshafiee/swephgo"
)

type SmokeResult struct {
	Version      string
	JulianDay    float64
	SunLongitude float64
	SunLatitude  float64
}

func Smoke() (SmokeResult, error) {
	version := make([]byte, 32)
	swephgo.Version(version)
	defer swephgo.Close()

	julianDay := swephgo.Julday(2026, 6, 7, 12, swephgo.SeGregCal)
	position := make([]float64, 6)
	errbuf := make([]byte, 256)

	flags := swephgo.SeflgSwieph | swephgo.SeflgSpeed
	if result := swephgo.CalcUt(julianDay, swephgo.SeSun, flags, position, errbuf); result < 0 {
		return SmokeResult{}, fmt.Errorf("swiss ephemeris calculation failed: %s", cString(errbuf))
	}

	return SmokeResult{
		Version:      cString(version),
		JulianDay:    julianDay,
		SunLongitude: position[0],
		SunLatitude:  position[1],
	}, nil
}

func cString(value []byte) string {
	return string(bytes.TrimRight(value, "\x00"))
}
