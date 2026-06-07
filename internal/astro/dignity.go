package astro

var exaltations = map[Planet]Sign{
	Sun:     Aries,
	Moon:    Taurus,
	Mercury: Virgo,
	Venus:   Pisces,
	Mars:    Capricorn,
	Jupiter: Cancer,
	Saturn:  Libra,
}

func EssentialStatus(planet Planet, sign Sign) string {
	if DomicileRuler(sign) == planet {
		return "domicile"
	}
	if exaltations[planet] == sign {
		return "exaltation"
	}
	opposite := Sign((int(sign) + 6) % 12)
	if DomicileRuler(opposite) == planet {
		return "detriment"
	}
	if exaltations[planet] == opposite {
		return "fall"
	}
	return "peregrine"
}
