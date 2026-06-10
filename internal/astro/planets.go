package astro

type Planet string

const (
	Sun                 Planet = "Sun"
	Moon                Planet = "Moon"
	Mercury             Planet = "Mercury"
	Venus               Planet = "Venus"
	Mars                Planet = "Mars"
	Jupiter             Planet = "Jupiter"
	Saturn              Planet = "Saturn"
	Uranus              Planet = "Uranus"
	Neptune             Planet = "Neptune"
	Pluto               Planet = "Pluto"
	NorthNode           Planet = "NorthNode"
	SouthNode           Planet = "SouthNode"
	Chiron              Planet = "Chiron"
	ParsFortunae        Planet = "ParsFortunae"
	MeanNorthNode       Planet = "MeanNorthNode"
	MeanSouthNode       Planet = "MeanSouthNode"
	BlackMoonLilith     Planet = "BlackMoonLilith"
	TrueBlackMoonLilith Planet = "TrueBlackMoonLilith"
	Earth               Planet = "Earth"
	Pholus              Planet = "Pholus"
	Ceres               Planet = "Ceres"
	Pallas              Planet = "Pallas"
	Juno                Planet = "Juno"
	Vesta               Planet = "Vesta"
	Varuna              Planet = "Varuna"
	Cupido              Planet = "Cupido"
	Hades               Planet = "Hades"
	Zeus                Planet = "Zeus"
	Kronos              Planet = "Kronos"
	Apollon             Planet = "Apollon"
	Admetos             Planet = "Admetos"
	Vulkanus            Planet = "Vulkanus"
	Poseidon            Planet = "Poseidon"
	Isis                Planet = "Isis"
	WhiteMoon           Planet = "WhiteMoon"
	Proserpina          Planet = "Proserpina"
)

var TraditionalPlanets = []Planet{Sun, Moon, Mercury, Venus, Mars, Jupiter, Saturn}
var ModernPlanets = []Planet{Sun, Moon, Mercury, Venus, Mars, Jupiter, Saturn, Uranus, Neptune, Pluto, NorthNode, SouthNode, Chiron, ParsFortunae}

var planetGlyphs = map[Planet]string{
	Sun:          "Q",
	Moon:         "W",
	Mercury:      "E",
	Venus:        "R",
	Mars:         "T",
	Jupiter:      "Y",
	Saturn:       "U",
	Uranus:       "I",
	Neptune:      "O",
	Pluto:        "P",
	NorthNode:    "\u008b",
	SouthNode:    "\u008c",
	Chiron:       "M",
	ParsFortunae: "<",
}

func (p Planet) Glyph() string {
	if glyph, ok := planetGlyphs[p]; ok {
		return glyph
	}
	return string(p)
}
