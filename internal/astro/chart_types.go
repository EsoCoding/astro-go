package astro

type ChartType string

const (
	ChartTypeNatal                ChartType = "natal"
	ChartTypeTransit              ChartType = "transit"
	ChartTypeSynastry             ChartType = "synastry"
	ChartTypeSecondaryProgression ChartType = "secondary_progression"
	ChartTypeSolarArc             ChartType = "solar_arc"
	ChartTypeSolarReturn          ChartType = "solar_return"
	ChartTypeLunarReturn          ChartType = "lunar_return"
)

func SupportedChartTypes() []ChartType {
	return []ChartType{
		ChartTypeNatal,
		ChartTypeTransit,
		ChartTypeSynastry,
		ChartTypeSecondaryProgression,
		ChartTypeSolarArc,
		ChartTypeSolarReturn,
		ChartTypeLunarReturn,
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
	case ChartTypeSolarArc:
		return "Solar Arc"
	case ChartTypeSolarReturn:
		return "Solar Return"
	case ChartTypeLunarReturn:
		return "Lunar Return"
	default:
		return string(t)
	}
}

func (t ChartType) RequiresComparisonChart() bool {
	return t == ChartTypeSynastry
}

func (t ChartType) RequiresReferenceTime() bool {
	switch t {
	case ChartTypeTransit, ChartTypeSecondaryProgression, ChartTypeSolarArc, ChartTypeSolarReturn, ChartTypeLunarReturn:
		return true
	default:
		return false
	}
}

func (t ChartType) SupportsDirectBirthData() bool {
	return t == ChartTypeNatal
}
