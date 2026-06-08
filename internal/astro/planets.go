package astro

type Planet string

const (
	Sun     Planet = "Sun"
	Moon    Planet = "Moon"
	Mercury Planet = "Mercury"
	Venus   Planet = "Venus"
	Mars    Planet = "Mars"
	Jupiter Planet = "Jupiter"
	Saturn  Planet = "Saturn"
)

var TraditionalPlanets = []Planet{Sun, Moon, Mercury, Venus, Mars, Jupiter, Saturn}

var planetGlyphs = map[Planet]string{
	Sun:     "Q",
	Moon:    "W",
	Mercury: "E",
	Venus:   "R",
	Mars:    "T",
	Jupiter: "Y",
	Saturn:  "U",
}

func (p Planet) Glyph() string {
	if glyph, ok := planetGlyphs[p]; ok {
		return glyph
	}
	return string(p)
}
