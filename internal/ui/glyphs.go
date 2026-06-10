package ui

import (
	"astro-go/internal/assets"
	"astro-go/internal/astro"

	"fyne.io/fyne/v2"
)

func astrologyFont() fyne.Resource {
	return assets.EnigmaAstrologyFont
}

var enigmaSignGlyphs = map[astro.Sign]string{
	astro.Aries:       "\uE000",
	astro.Taurus:      "\uE001",
	astro.Gemini:      "\uE002",
	astro.Cancer:      "\uE003",
	astro.Leo:         "\uE004",
	astro.Virgo:       "\uE005",
	astro.Libra:       "\uE006",
	astro.Scorpio:     "\uE007",
	astro.Sagittarius: "\uE008",
	astro.Capricorn:   "\uE009",
	astro.Aquarius:    "\uE010",
	astro.Pisces:      "\uE011",
}

var enigmaPlanetGlyphs = map[astro.Planet]string{
	astro.Sun:                 "\uE200",
	astro.Moon:                "\uE201",
	astro.Mercury:             "\uE202",
	astro.Venus:               "\uE203",
	astro.Earth:               "\uE204",
	astro.Mars:                "\uE205",
	astro.Jupiter:             "\uE206",
	astro.Saturn:              "\uE207",
	astro.Uranus:              "\uE208",
	astro.Neptune:             "\uE209",
	astro.Pluto:               "\uE210",
	astro.Chiron:              "\uE400",
	astro.Pholus:              "\uE402",
	astro.Varuna:              "\uE403",
	astro.Ceres:               "\uE411",
	astro.Pallas:              "\uE412",
	astro.Juno:                "\uE413",
	astro.Vesta:               "\uE414",
	astro.ParsFortunae:        "\uF400",
	astro.NorthNode:           "\uE525",
	astro.SouthNode:           "\uE526",
	astro.MeanNorthNode:       "\uE523",
	astro.MeanSouthNode:       "\uE524",
	astro.BlackMoonLilith:     "\uE530",
	astro.TrueBlackMoonLilith: "\uE531",
	astro.Cupido:              "\uE600",
	astro.Hades:               "\uE601",
	astro.Zeus:                "\uE602",
	astro.Kronos:              "\uE603",
	astro.Apollon:             "\uE604",
	astro.Admetos:             "\uE605",
	astro.Vulkanus:            "\uE606",
	astro.Poseidon:            "\uE607",
	astro.Isis:                "\uE611",
	astro.Proserpina:          "\uE616",
	astro.WhiteMoon:           "\uE532",
}

func signGlyph(sign astro.Sign) string {
	if glyph, ok := enigmaSignGlyphs[sign]; ok {
		return glyph
	}
	return "?"
}

func planetGlyph(planet astro.Planet) string {
	if glyph, ok := enigmaPlanetGlyphs[planet]; ok {
		return glyph
	}
	return "?"
}

func aspectGlyph(typ astro.AspectType) string {
	switch typ {
	case astro.Conjunction:
		return "\uE700"
	case astro.Sextile:
		return "\uE740"
	case astro.Square:
		return "\uE730"
	case astro.Trine:
		return "\uE720"
	case astro.Opposition:
		return "\uE710"
	default:
		return "?"
	}
}
