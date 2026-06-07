package astro

import "fmt"

type Sign int

const (
	Aries Sign = iota
	Taurus
	Gemini
	Cancer
	Leo
	Virgo
	Libra
	Scorpio
	Sagittarius
	Capricorn
	Aquarius
	Pisces
)

var signNames = [...]string{
	"Aries",
	"Taurus",
	"Gemini",
	"Cancer",
	"Leo",
	"Virgo",
	"Libra",
	"Scorpio",
	"Sagittarius",
	"Capricorn",
	"Aquarius",
	"Pisces",
}

var signGlyphs = [...]string{"Ar", "Ta", "Ge", "Cn", "Le", "Vi", "Li", "Sc", "Sg", "Cp", "Aq", "Pi"}

func NormalizeDegrees(degrees float64) float64 {
	for degrees < 0 {
		degrees += 360
	}
	for degrees >= 360 {
		degrees -= 360
	}
	return degrees
}

func SignFromLongitude(longitude float64) Sign {
	return Sign(int(NormalizeDegrees(longitude) / 30))
}

func DegreeInSign(longitude float64) float64 {
	return NormalizeDegrees(longitude) - float64(SignFromLongitude(longitude))*30
}

func (s Sign) String() string {
	if s < Aries || s > Pisces {
		return fmt.Sprintf("Sign(%d)", s)
	}
	return signNames[s]
}

func (s Sign) Glyph() string {
	if s < Aries || s > Pisces {
		return "?"
	}
	return signGlyphs[s]
}
