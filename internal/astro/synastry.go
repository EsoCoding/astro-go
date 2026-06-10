package astro

type InterAspect struct {
	Inner Planet
	Outer Planet
	Type  AspectType
	Orb   float64
	Exact float64
}

type SynastryChart struct {
	Name         string
	InnerChart   Chart
	OuterChart   Chart
	InterAspects []InterAspect
}

func TraditionalInterAspects(inner []PlanetPosition, outer []PlanetPosition) []InterAspect {
	inner = AspectablePositions(inner)
	outer = AspectablePositions(outer)
	var aspects []InterAspect
	for i := 0; i < len(inner); i++ {
		for j := 0; j < len(outer); j++ {
			distance := angularDistance(inner[i].Longitude, outer[j].Longitude)
			for _, rule := range traditionalAspectRules {
				orb := absFloat(distance - rule.exact)
				if orb <= rule.orb {
					aspects = append(aspects, InterAspect{
						Inner: inner[i].Planet,
						Outer: outer[j].Planet,
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

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
