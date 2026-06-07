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
	Sun:     "Su",
	Moon:    "Mo",
	Mercury: "Me",
	Venus:   "Ve",
	Mars:    "Ma",
	Jupiter: "Ju",
	Saturn:  "Sa",
}

func (p Planet) Glyph() string {
	if glyph, ok := planetGlyphs[p]; ok {
		return glyph
	}
	return string(p)
}
