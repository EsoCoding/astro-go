package ui

import (
	"fmt"

	"astro-go/internal/astro"
)

func formatZodiac(longitude float64) string {
	sign := astro.SignFromLongitude(longitude)
	return fmt.Sprintf("%05.2f %s", astro.DegreeInSign(longitude), sign)
}

func retrogradeMarker(retrograde bool) string {
	if retrograde {
		return " R"
	}
	return ""
}
