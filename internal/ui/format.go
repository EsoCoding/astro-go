package ui

import (
	"fmt"
	"math"
	"strings"

	"astro-go/internal/astro"
)

func formatZodiacDMS(longitude float64) string {
	sign := astro.SignFromLongitude(longitude)
	degreeInSign := astro.DegreeInSign(longitude)
	degrees := int(math.Floor(degreeInSign))
	minuteFloat := (degreeInSign - float64(degrees)) * 60
	minutes := int(math.Floor(minuteFloat))
	seconds := int(math.Round((minuteFloat - float64(minutes)) * 60))
	if seconds == 60 {
		seconds = 0
		minutes++
	}
	if minutes == 60 {
		minutes = 0
		degrees++
	}
	if degrees == 30 {
		degrees = 29
		minutes = 59
		seconds = 59
	}
	return fmt.Sprintf("%02d°%02d'%02d\" %s", degrees, minutes, seconds, sign)
}

func retrogradeMarker(retrograde bool) string {
	if retrograde {
		return " R"
	}
	return ""
}

func formatCoordsDMS(lat, lon float64) string {
	latDir := "N"
	if lat < 0 {
		latDir = "S"
		lat = -lat
	}
	lonDir := "E"
	if lon < 0 {
		lonDir = "W"
		lon = -lon
	}

	formatPart := func(val float64, dir string, isLon bool) string {
		deg := int(math.Floor(val))
		minFloat := (val - float64(deg)) * 60
		min := int(math.Floor(minFloat))
		sec := int(math.Round((minFloat - float64(min)) * 60))
		if sec == 60 {
			sec = 0
			min++
		}
		if min == 60 {
			min = 0
			deg++
		}

		degWidth := 2
		if isLon {
			degWidth = 3
		}

		return fmt.Sprintf("%0*d°%s%02d'%02d\"", degWidth, deg, dir, min, sec)
	}

	return fmt.Sprintf("%s %s", formatPart(lat, latDir, false), formatPart(lon, lonDir, true))
}

func shortenLocationName(location string) string {
	parts := strings.Split(location, ",")
	if len(parts) >= 3 {
		return strings.TrimSpace(parts[0]) + ", " + strings.TrimSpace(parts[len(parts)-1])
	}
	return location
}
