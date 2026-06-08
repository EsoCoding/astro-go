package astro

type HouseSystem string

const (
	HouseSystemPlacidus             HouseSystem = "P"
	HouseSystemKoch                 HouseSystem = "K"
	HouseSystemPorphyry             HouseSystem = "O"
	HouseSystemRegiomontanus        HouseSystem = "R"
	HouseSystemCampanus             HouseSystem = "C"
	HouseSystemEqual                HouseSystem = "A"
	HouseSystemEqualMC              HouseSystem = "E"
	HouseSystemVehlowEqual          HouseSystem = "V"
	HouseSystemWholeSign            HouseSystem = "W"
	HouseSystemAlcabitius           HouseSystem = "B"
	HouseSystemTopocentric          HouseSystem = "T"
	HouseSystemMorinus              HouseSystem = "M"
	HouseSystemKrusinskiPisaGoelzer HouseSystem = "U"
	HouseSystemHorizonAzimuth       HouseSystem = "H"
	HouseSystemMeridian             HouseSystem = "X"
	HouseSystemAPC                  HouseSystem = "Y"
	HouseSystemGauquelin            HouseSystem = "G"
)

func SupportedHouseSystems() []HouseSystem {
	return []HouseSystem{
		HouseSystemPlacidus,
		HouseSystemKoch,
		HouseSystemPorphyry,
		HouseSystemRegiomontanus,
		HouseSystemCampanus,
		HouseSystemEqual,
		HouseSystemEqualMC,
		HouseSystemVehlowEqual,
		HouseSystemWholeSign,
		HouseSystemAlcabitius,
		HouseSystemTopocentric,
		HouseSystemMorinus,
		HouseSystemKrusinskiPisaGoelzer,
		HouseSystemHorizonAzimuth,
		HouseSystemMeridian,
		HouseSystemAPC,
		HouseSystemGauquelin,
	}
}

func DefaultHouseSystem() HouseSystem {
	return HouseSystemWholeSign
}

func HouseSystemFromCode(value string) HouseSystem {
	system := HouseSystem(value)
	for _, candidate := range SupportedHouseSystems() {
		if candidate == system {
			return candidate
		}
	}
	return DefaultHouseSystem()
}

func (h HouseSystem) Code() int {
	if h == "" {
		h = DefaultHouseSystem()
	}
	return int(h[0])
}

func (h HouseSystem) Label() string {
	switch h {
	case HouseSystemPlacidus:
		return "Placidus"
	case HouseSystemKoch:
		return "Koch"
	case HouseSystemPorphyry:
		return "Porphyry"
	case HouseSystemRegiomontanus:
		return "Regiomontanus"
	case HouseSystemCampanus:
		return "Campanus"
	case HouseSystemEqual:
		return "Equal"
	case HouseSystemEqualMC:
		return "Equal (MC)"
	case HouseSystemVehlowEqual:
		return "Vehlow Equal"
	case HouseSystemWholeSign:
		return "Whole Sign"
	case HouseSystemAlcabitius:
		return "Alcabitius"
	case HouseSystemTopocentric:
		return "Topocentric"
	case HouseSystemMorinus:
		return "Morinus"
	case HouseSystemKrusinskiPisaGoelzer:
		return "Krusinski-Pisa-Goelzer"
	case HouseSystemHorizonAzimuth:
		return "Horizon / Azimuth"
	case HouseSystemMeridian:
		return "Meridian"
	case HouseSystemAPC:
		return "APC"
	case HouseSystemGauquelin:
		return "Gauquelin Sectors"
	default:
		return string(h)
	}
}

func HouseSystemOptions() []string {
	options := make([]string, 0, len(SupportedHouseSystems()))
	for _, houseSystem := range SupportedHouseSystems() {
		options = append(options, houseSystem.Label())
	}
	return options
}

func HouseSystemFromLabel(value string) HouseSystem {
	for _, houseSystem := range SupportedHouseSystems() {
		if houseSystem.Label() == value {
			return houseSystem
		}
	}
	return DefaultHouseSystem()
}

func (h HouseSystem) HouseCount() int {
	if h == HouseSystemGauquelin {
		return 36
	}
	return 12
}
