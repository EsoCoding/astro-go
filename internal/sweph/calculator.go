package sweph

import (
	"fmt"

	"astro-go/internal/astro"

	"github.com/mshafiee/swephgo"
)

type Calculator struct{}

func NewCalculator() Calculator {
	return Calculator{}
}

func (Calculator) NatalChart(data astro.BirthData) (astro.Chart, error) {
	hour := float64(data.DateTimeUTC.Hour()) +
		float64(data.DateTimeUTC.Minute())/60 +
		float64(data.DateTimeUTC.Second())/3600
	julianDay := swephgo.Julday(
		data.DateTimeUTC.Year(),
		int(data.DateTimeUTC.Month()),
		data.DateTimeUTC.Day(),
		hour,
		swephgo.SeGregCal,
	)

	asc, mc, err := angles(julianDay, data.LatitudeDegrees, data.LongitudeDegrees)
	if err != nil {
		return astro.Chart{}, err
	}

	ascSign := astro.SignFromLongitude(asc)
	houses := astro.WholeSignHouses(ascSign)
	planets, err := planetPositions(julianDay, ascSign)
	if err != nil {
		return astro.Chart{}, err
	}

	return astro.Chart{
		Name:        data.Name,
		DateTimeUTC: data.DateTimeUTC,
		Latitude:    data.LatitudeDegrees,
		Longitude:   data.LongitudeDegrees,
		JulianDay:   julianDay,
		Ascendant:   angle("Asc", asc),
		MC:          angle("MC", mc),
		Houses:      houses,
		Planets:     planets,
		Aspects:     astro.TraditionalAspects(planets),
	}, nil
}

func angles(julianDay, latitude, longitude float64) (float64, float64, error) {
	cusps := make([]float64, 13)
	ascmc := make([]float64, 10)
	if result := swephgo.Houses(julianDay, latitude, longitude, int('W'), cusps, ascmc); result < 0 {
		return 0, 0, fmt.Errorf("failed to calculate chart angles")
	}
	return astro.NormalizeDegrees(ascmc[0]), astro.NormalizeDegrees(ascmc[1]), nil
}

func planetPositions(julianDay float64, ascendant astro.Sign) ([]astro.PlanetPosition, error) {
	swephIDs := map[astro.Planet]int{
		astro.Sun:     swephgo.SeSun,
		astro.Moon:    swephgo.SeMoon,
		astro.Mercury: swephgo.SeMercury,
		astro.Venus:   swephgo.SeVenus,
		astro.Mars:    swephgo.SeMars,
		astro.Jupiter: swephgo.SeJupiter,
		astro.Saturn:  swephgo.SeSaturn,
	}

	flags := swephgo.SeflgMoseph | swephgo.SeflgSpeed
	positions := make([]astro.PlanetPosition, 0, len(astro.TraditionalPlanets))
	for _, planet := range astro.TraditionalPlanets {
		values := make([]float64, 6)
		errbuf := make([]byte, 256)
		if result := swephgo.CalcUt(julianDay, swephIDs[planet], flags, values, errbuf); result < 0 {
			return nil, fmt.Errorf("failed to calculate %s: %s", planet, cString(errbuf))
		}

		longitude := astro.NormalizeDegrees(values[0])
		sign := astro.SignFromLongitude(longitude)
		positions = append(positions, astro.PlanetPosition{
			Planet:          planet,
			Longitude:       longitude,
			Latitude:        values[1],
			Speed:           values[3],
			Sign:            sign,
			DegreeInSign:    astro.DegreeInSign(longitude),
			WholeSignHouse:  astro.WholeSignHouse(ascendant, longitude),
			Retrograde:      values[3] < 0,
			DomicileRuler:   astro.DomicileRuler(sign),
			EssentialStatus: astro.EssentialStatus(planet, sign),
		})
	}
	return positions, nil
}

func angle(name string, longitude float64) astro.Angle {
	return astro.Angle{
		Name:         name,
		Longitude:    longitude,
		Sign:         astro.SignFromLongitude(longitude),
		DegreeInSign: astro.DegreeInSign(longitude),
	}
}
