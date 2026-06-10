package astro

type ChartType string

const (
	ChartTypeNatal                ChartType = "natal"
	ChartTypeTransit              ChartType = "transit"
	ChartTypeSynastry             ChartType = "synastry"
	ChartTypeSecondaryProgression ChartType = "secondary_progression"
	ChartTypeTertiaryProgression  ChartType = "tertiary_progression"
	ChartTypeSolarArc             ChartType = "solar_arc"
	ChartTypeSolarReturn          ChartType = "solar_return"
	ChartTypeLunarReturn          ChartType = "lunar_return"
	ChartTypePrimaryDirection     ChartType = "primary_direction"
	ChartTypeComposite            ChartType = "composite"
	ChartTypeDavison              ChartType = "davison"
)

func SupportedChartTypes() []ChartType {
	return []ChartType{
		ChartTypeNatal,
		ChartTypeTransit,
		ChartTypeSynastry,
		ChartTypeSecondaryProgression,
		ChartTypeTertiaryProgression,
		ChartTypeSolarArc,
		ChartTypeSolarReturn,
		ChartTypeLunarReturn,
		ChartTypePrimaryDirection,
		ChartTypeComposite,
		ChartTypeDavison,
	}
}

func (t ChartType) String() string {
	switch t {
	case ChartTypeNatal:
		return "Natal"
	case ChartTypeTransit:
		return "Transit"
	case ChartTypeSynastry:
		return "Synastry"
	case ChartTypeSecondaryProgression:
		return "Secondary Progression"
	case ChartTypeTertiaryProgression:
		return "Tertiary Progression"
	case ChartTypeSolarArc:
		return "Solar Arc"
	case ChartTypeSolarReturn:
		return "Solar Return"
	case ChartTypeLunarReturn:
		return "Lunar Return"
	case ChartTypePrimaryDirection:
		return "Primary Directions"
	case ChartTypeComposite:
		return "Composite"
	case ChartTypeDavison:
		return "Davison"
	default:
		return string(t)
	}
}

func (t ChartType) RequiresComparisonChart() bool {
	return t == ChartTypeSynastry || t == ChartTypeComposite || t == ChartTypeDavison
}

func (t ChartType) RequiresReferenceTime() bool {
	switch t {
	case ChartTypeTransit, ChartTypeSecondaryProgression, ChartTypeTertiaryProgression, ChartTypeSolarArc, ChartTypeSolarReturn, ChartTypeLunarReturn, ChartTypePrimaryDirection:
		return true
	default:
		return false
	}
}

func (t ChartType) SupportsDirectBirthData() bool {
	return t == ChartTypeNatal
}
