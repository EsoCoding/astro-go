package astro

var domicileRulers = map[Sign]Planet{
	Aries:       Mars,
	Taurus:      Venus,
	Gemini:      Mercury,
	Cancer:      Moon,
	Leo:         Sun,
	Virgo:       Mercury,
	Libra:       Venus,
	Scorpio:     Mars,
	Sagittarius: Jupiter,
	Capricorn:   Saturn,
	Aquarius:    Saturn,
	Pisces:      Jupiter,
}

func DomicileRuler(sign Sign) Planet {
	return domicileRulers[sign]
}

func WholeSignHouses(ascendant Sign) []House {
	houses := make([]House, 12)
	for i := range houses {
		longitude := NormalizeDegrees(float64((int(ascendant) + i) * 30))
		sign := SignFromLongitude(longitude)
		houses[i] = House{
			Number:        i + 1,
			CuspLongitude: longitude,
			Sign:          sign,
			Ruler:         DomicileRuler(sign),
		}
	}
	return houses
}

func WholeSignHouse(ascendant Sign, longitude float64) int {
	sign := SignFromLongitude(longitude)
	return ((int(sign)-int(ascendant)+12)%12 + 1)
}

func HousesFromCusps(cusps []float64) []House {
	houses := make([]House, 0, len(cusps))
	for i, cusp := range cusps {
		sign := SignFromLongitude(cusp)
		houses = append(houses, House{
			Number:        i + 1,
			CuspLongitude: NormalizeDegrees(cusp),
			Sign:          sign,
			Ruler:         DomicileRuler(sign),
		})
	}
	return houses
}
