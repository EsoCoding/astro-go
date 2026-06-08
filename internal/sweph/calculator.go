package sweph

import (
	"fmt"
	"math"

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

	houseSystem := data.HouseSystem
	if houseSystem == "" {
		houseSystem = astro.DefaultHouseSystem()
	}

	houseCusps, asc, mc, armc, err := housesAndAngles(julianDay, data.LatitudeDegrees, data.LongitudeDegrees, houseSystem)
	if err != nil {
		return astro.Chart{}, err
	}

	houses := astro.HousesFromCusps(houseCusps)
	eps, err := obliquity(julianDay)
	if err != nil {
		return astro.Chart{}, err
	}
	planets, err := planetPositions(julianDay, data.LatitudeDegrees, armc, eps, houseSystem)
	if err != nil {
		return astro.Chart{}, err
	}

	return astro.Chart{
		Name:         data.Name,
		DateTimeUTC:  data.DateTimeUTC,
		LocationName: data.LocationName,
		Latitude:     data.LatitudeDegrees,
		Longitude:    data.LongitudeDegrees,
		HouseSystem:  houseSystem,
		JulianDay:    julianDay,
		Ascendant:    angle("Asc", asc),
		MC:           angle("MC", mc),
		Houses:       houses,
		Planets:      planets,
		Aspects:      astro.TraditionalAspects(planets),
		UTCOffset:    data.UTCOffset,
		TimezoneName: data.TimezoneName,
		ChartType:    data.ChartType,
	}, nil
}

func housesAndAngles(julianDay, latitude, longitude float64, houseSystem astro.HouseSystem) ([]float64, float64, float64, float64, error) {
	cusps := make([]float64, houseSystem.HouseCount()+1)
	ascmc := make([]float64, 10)
	if result := swephgo.Houses(julianDay, latitude, longitude, houseSystem.Code(), cusps, ascmc); result < 0 {
		return nil, 0, 0, 0, fmt.Errorf("failed to calculate houses for %s", houseSystem.Label())
	}
	values := make([]float64, 0, len(cusps)-1)
	for i := 1; i < len(cusps); i++ {
		values = append(values, astro.NormalizeDegrees(cusps[i]))
	}
	return values, astro.NormalizeDegrees(ascmc[0]), astro.NormalizeDegrees(ascmc[1]), astro.NormalizeDegrees(ascmc[2]), nil
}

func obliquity(julianDay float64) (float64, error) {
	values := make([]float64, 6)
	errbuf := make([]byte, 256)
	if result := swephgo.CalcUt(julianDay, swephgo.SeEclNut, 0, values, errbuf); result < 0 {
		return 0, fmt.Errorf("failed to calculate obliquity: %s", cString(errbuf))
	}
	return values[0], nil
}

func planetPositions(julianDay, latitude, armc, eps float64, houseSystem astro.HouseSystem) ([]astro.PlanetPosition, error) {
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
		house, err := planetHouse(values[0], values[1], latitude, armc, eps, houseSystem)
		if err != nil {
			return nil, err
		}
		positions = append(positions, astro.PlanetPosition{
			Planet:          planet,
			Longitude:       longitude,
			Latitude:        values[1],
			Speed:           values[3],
			Sign:            sign,
			DegreeInSign:    astro.DegreeInSign(longitude),
			House:           house,
			Retrograde:      values[3] < 0,
			DomicileRuler:   astro.DomicileRuler(sign),
			EssentialStatus: astro.EssentialStatus(planet, sign),
		})
	}
	return positions, nil
}

func planetHouse(longitude, latitude, geolat, armc, eps float64, houseSystem astro.HouseSystem) (int, error) {
	errbuf := make([]byte, 256)
	xpin := []float64{longitude, latitude}
	position := swephgo.HousePos(armc, geolat, eps, houseSystem.Code(), xpin, errbuf)
	if math.IsNaN(position) || position <= 0 {
		return 0, fmt.Errorf("failed to calculate house position: %s", cString(errbuf))
	}
	return int(math.Ceil(position - 1e-9)), nil
}

func angle(name string, longitude float64) astro.Angle {
	return astro.Angle{
		Name:         name,
		Longitude:    longitude,
		Sign:         astro.SignFromLongitude(longitude),
		DegreeInSign: astro.DegreeInSign(longitude),
	}
}
