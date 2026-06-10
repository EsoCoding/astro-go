package astro

type ChartObjectCategory string

const (
	ChartObjectCategoryTraditional ChartObjectCategory = "Traditional"
	ChartObjectCategoryModern      ChartObjectCategory = "Modern"
	ChartObjectCategoryNodes       ChartObjectCategory = "Nodes"
	ChartObjectCategoryLots        ChartObjectCategory = "Lots"
	ChartObjectCategoryAsteroids   ChartObjectCategory = "Asteroids"
	ChartObjectCategoryFictitious  ChartObjectCategory = "Fictitious"
)

type ChartObjectSpec struct {
	Planet           Planet
	Name             string
	Category         ChartObjectCategory
	DefaultEnabled   bool
	IncludeInAspects bool
	HasDignity       bool
}

var chartObjectCatalog = []ChartObjectSpec{
	{Sun, "Sun", ChartObjectCategoryTraditional, true, true, true},
	{Moon, "Moon", ChartObjectCategoryTraditional, true, true, true},
	{Mercury, "Mercury", ChartObjectCategoryTraditional, true, true, true},
	{Venus, "Venus", ChartObjectCategoryTraditional, true, true, true},
	{Mars, "Mars", ChartObjectCategoryTraditional, true, true, true},
	{Jupiter, "Jupiter", ChartObjectCategoryTraditional, true, true, true},
	{Saturn, "Saturn", ChartObjectCategoryTraditional, true, true, true},
	{Uranus, "Uranus", ChartObjectCategoryModern, true, true, false},
	{Neptune, "Neptune", ChartObjectCategoryModern, true, true, false},
	{Pluto, "Pluto", ChartObjectCategoryModern, true, true, false},
	{NorthNode, "North Node", ChartObjectCategoryNodes, true, true, false},
	{SouthNode, "South Node", ChartObjectCategoryNodes, true, true, false},
	{Chiron, "Chiron", ChartObjectCategoryAsteroids, true, true, false},
	{ParsFortunae, "Part of Fortune", ChartObjectCategoryLots, true, true, false},
	{MeanNorthNode, "Mean North Node", ChartObjectCategoryNodes, false, false, false},
	{MeanSouthNode, "Mean South Node", ChartObjectCategoryNodes, false, false, false},
	{BlackMoonLilith, "Black Moon Lilith", ChartObjectCategoryNodes, false, false, false},
	{TrueBlackMoonLilith, "True Black Moon Lilith", ChartObjectCategoryNodes, false, false, false},
	{Earth, "Earth", ChartObjectCategoryModern, false, false, false},
	{Pholus, "Pholus", ChartObjectCategoryAsteroids, false, false, false},
	{Ceres, "Ceres", ChartObjectCategoryAsteroids, false, false, false},
	{Pallas, "Pallas Athene", ChartObjectCategoryAsteroids, false, false, false},
	{Juno, "Juno", ChartObjectCategoryAsteroids, false, false, false},
	{Vesta, "Vesta", ChartObjectCategoryAsteroids, false, false, false},
	{Varuna, "Varuna", ChartObjectCategoryAsteroids, false, false, false},
	{Cupido, "Cupido", ChartObjectCategoryFictitious, false, false, false},
	{Hades, "Hades", ChartObjectCategoryFictitious, false, false, false},
	{Zeus, "Zeus", ChartObjectCategoryFictitious, false, false, false},
	{Kronos, "Kronos", ChartObjectCategoryFictitious, false, false, false},
	{Apollon, "Apollon", ChartObjectCategoryFictitious, false, false, false},
	{Admetos, "Admetos", ChartObjectCategoryFictitious, false, false, false},
	{Vulkanus, "Vulkanus", ChartObjectCategoryFictitious, false, false, false},
	{Poseidon, "Poseidon", ChartObjectCategoryFictitious, false, false, false},
	{Isis, "TransPluto / Isis", ChartObjectCategoryFictitious, false, false, false},
	{WhiteMoon, "White Moon Selena", ChartObjectCategoryFictitious, false, false, false},
	{Proserpina, "Proserpina", ChartObjectCategoryFictitious, false, false, false},
}

func ChartObjectCatalog() []ChartObjectSpec {
	return append([]ChartObjectSpec(nil), chartObjectCatalog...)
}

func DefaultEnabledChartObjects() []Planet {
	objects := make([]Planet, 0, len(chartObjectCatalog))
	for _, spec := range chartObjectCatalog {
		if spec.DefaultEnabled {
			objects = append(objects, spec.Planet)
		}
	}
	return objects
}

func ChartObjectSpecFor(planet Planet) (ChartObjectSpec, bool) {
	for _, spec := range chartObjectCatalog {
		if spec.Planet == planet {
			return spec, true
		}
	}
	return ChartObjectSpec{}, false
}

func EnabledChartObjectSet(objects []Planet) map[Planet]bool {
	if objects == nil {
		objects = DefaultEnabledChartObjects()
	}
	enabled := make(map[Planet]bool, len(objects))
	for _, object := range objects {
		enabled[object] = true
	}
	return enabled
}

func AspectablePositions(positions []PlanetPosition) []PlanetPosition {
	filtered := make([]PlanetPosition, 0, len(positions))
	for _, position := range positions {
		spec, ok := ChartObjectSpecFor(position.Planet)
		if ok && !spec.IncludeInAspects {
			continue
		}
		filtered = append(filtered, position)
	}
	return filtered
}
