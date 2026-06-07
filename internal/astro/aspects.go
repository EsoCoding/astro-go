package astro

import "math"

type AspectType string

const (
	Conjunction AspectType = "conjunction"
	Sextile     AspectType = "sextile"
	Square      AspectType = "square"
	Trine       AspectType = "trine"
	Opposition  AspectType = "opposition"
)

type aspectRule struct {
	typ   AspectType
	exact float64
	orb   float64
}

var traditionalAspectRules = []aspectRule{
	{Conjunction, 0, 8},
	{Sextile, 60, 6},
	{Square, 90, 7},
	{Trine, 120, 7},
	{Opposition, 180, 8},
}

func TraditionalAspects(positions []PlanetPosition) []Aspect {
	var aspects []Aspect
	for i := 0; i < len(positions); i++ {
		for j := i + 1; j < len(positions); j++ {
			distance := angularDistance(positions[i].Longitude, positions[j].Longitude)
			for _, rule := range traditionalAspectRules {
				orb := math.Abs(distance - rule.exact)
				if orb <= rule.orb {
					aspects = append(aspects, Aspect{
						From:  positions[i].Planet,
						To:    positions[j].Planet,
						Type:  rule.typ,
						Orb:   orb,
						Exact: rule.exact,
					})
					break
				}
			}
		}
	}
	return aspects
}

func angularDistance(a, b float64) float64 {
	distance := math.Abs(NormalizeDegrees(a) - NormalizeDegrees(b))
	if distance > 180 {
		return 360 - distance
	}
	return distance
}
