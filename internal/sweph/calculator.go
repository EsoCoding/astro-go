package sweph

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"

	"astro-go/internal/astro"

	swisseph "github.com/tejzpr/go-swisseph"
)

type Calculator struct{}

var configureEphemerisPathOnce sync.Once

func NewCalculator() Calculator {
	return Calculator{}
}

func (Calculator) NatalChart(data astro.BirthData) (astro.Chart, error) {
	configureEphemerisPath()

	hour := float64(data.DateTimeUTC.Hour()) +
		float64(data.DateTimeUTC.Minute())/60 +
		float64(data.DateTimeUTC.Second())/3600
	julianDay := swisseph.Julday(
		int32(data.DateTimeUTC.Year()),
		int32(data.DateTimeUTC.Month()),
		int32(data.DateTimeUTC.Day()),
		hour,
		swisseph.GregCal,
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
	planets, err := planetPositions(julianDay, data.LatitudeDegrees, asc, armc, eps, houseSystem)
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
	result := swisseph.Houses(julianDay, latitude, longitude, byte(houseSystem.Code()))
	if result.Flag < 0 {
		return nil, 0, 0, 0, fmt.Errorf("failed to calculate houses for %s", houseSystem.Label())
	}
	values := make([]float64, 0, houseSystem.HouseCount())
	for _, cusp := range result.Houses {
		values = append(values, astro.NormalizeDegrees(cusp))
	}
	return values, astro.NormalizeDegrees(result.Points[swisseph.Asc]), astro.NormalizeDegrees(result.Points[swisseph.MC]), astro.NormalizeDegrees(result.Points[swisseph.ARMC]), nil
}

func configureEphemerisPath() {
	configureEphemerisPathOnce.Do(func() {
		for _, candidate := range ephemerisPathCandidates() {
			if candidate == "" {
				continue
			}
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				swisseph.SetEphePath(candidate)
				return
			}
		}
	})
}

func ephemerisPathCandidates() []string {
	candidates := []string{os.Getenv("ASTRO_GO_EPHE_PATH")}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "third_party", "swisseph", "ephe"))
	}
	if executable, err := os.Executable(); err == nil {
		dir := filepath.Dir(executable)
		candidates = append(candidates,
			filepath.Join(dir, "third_party", "swisseph", "ephe"),
			filepath.Join(dir, "..", "third_party", "swisseph", "ephe"),
		)
	}
	return candidates
}

func obliquity(julianDay float64) (float64, error) {
	result := swisseph.CalcUT(julianDay, swisseph.EclNut, 0)
	if result.Flag < 0 {
		return 0, fmt.Errorf("failed to calculate obliquity: %s", result.Error)
	}
	return result.Data[0], nil
}

func planetPositions(julianDay, latitude, asc, armc, eps float64, houseSystem astro.HouseSystem) ([]astro.PlanetPosition, error) {
	swephIDs := map[astro.Planet]int32{
		astro.Sun:       swisseph.Sun,
		astro.Moon:      swisseph.Moon,
		astro.Mercury:   swisseph.Mercury,
		astro.Venus:     swisseph.Venus,
		astro.Mars:      swisseph.Mars,
		astro.Jupiter:   swisseph.Jupiter,
		astro.Saturn:    swisseph.Saturn,
		astro.Uranus:    swisseph.Uranus,
		astro.Neptune:   swisseph.Neptune,
		astro.Pluto:     swisseph.Pluto,
		astro.NorthNode: swisseph.TrueNode,
		astro.Chiron:    swisseph.Chiron,
	}

	flags := int32(swisseph.FlagSwieph | swisseph.FlagSpeed)
	positions := make([]astro.PlanetPosition, 0, len(astro.ModernPlanets))

	var sunLong, moonLong float64

	for _, planet := range astro.ModernPlanets {
		if planet == astro.SouthNode || planet == astro.ParsFortunae {
			continue // Handled manually below
		}
		result := swisseph.CalcUT(julianDay, swephIDs[planet], flags)
		if result.Flag < 0 {
			// If it's a traditional planet, fail. Otherwise, skip gracefully.
			isTraditional := false
			for _, tp := range astro.TraditionalPlanets {
				if planet == tp {
					isTraditional = true
					break
				}
			}
			if isTraditional {
				return nil, fmt.Errorf("failed to calculate %s: %s", planet, result.Error)
			}
			continue
		}

		values := result.Data
		longitude := astro.NormalizeDegrees(values[0])

		if planet == astro.Sun {
			sunLong = longitude
		} else if planet == astro.Moon {
			moonLong = longitude
		}

		sign := astro.SignFromLongitude(longitude)
		house, err := planetHouse(longitude, values[1], latitude, armc, eps, houseSystem)
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

	// Calculate South Node
	var nnLong float64
	for _, p := range positions {
		if p.Planet == astro.NorthNode {
			nnLong = p.Longitude
			break
		}
	}
	snLong := astro.NormalizeDegrees(nnLong + 180.0)
	snSign := astro.SignFromLongitude(snLong)
	snHouse, _ := planetHouse(snLong, 0, latitude, armc, eps, houseSystem) // 0 lat approx

	positions = append(positions, astro.PlanetPosition{
		Planet:          astro.SouthNode,
		Longitude:       snLong,
		Latitude:        0,
		Speed:           0, // Approximate
		Sign:            snSign,
		DegreeInSign:    astro.DegreeInSign(snLong),
		House:           snHouse,
		Retrograde:      true, // Nodes are generally retrograde
		DomicileRuler:   astro.DomicileRuler(snSign),
		EssentialStatus: astro.EssentialStatus(astro.SouthNode, snSign),
	})

	// Calculate Part of Fortune using the same horizon test as astro-server-2:
	// Sun 180-360 degrees counter-clockwise from ASC is above the horizon.
	isDayChart := astro.NormalizeDegrees(sunLong-asc) >= 180

	var pfLong float64
	if isDayChart {
		pfLong = astro.NormalizeDegrees(asc + moonLong - sunLong)
	} else {
		pfLong = astro.NormalizeDegrees(asc + sunLong - moonLong)
	}
	pfSign := astro.SignFromLongitude(pfLong)
	pfHouse, _ := planetHouse(pfLong, 0, latitude, armc, eps, houseSystem) // 0 lat approx

	positions = append(positions, astro.PlanetPosition{
		Planet:          astro.ParsFortunae,
		Longitude:       pfLong,
		Latitude:        0,
		Speed:           0,
		Sign:            pfSign,
		DegreeInSign:    astro.DegreeInSign(pfLong),
		House:           pfHouse,
		Retrograde:      false,
		DomicileRuler:   astro.DomicileRuler(pfSign),
		EssentialStatus: astro.EssentialStatus(astro.ParsFortunae, pfSign),
	})

	return positions, nil
}

func planetHouse(longitude, latitude, geolat, armc, eps float64, houseSystem astro.HouseSystem) (int, error) {
	position, err := swisseph.HousePos(armc, geolat, eps, byte(houseSystem.Code()), longitude, latitude)
	if math.IsNaN(position) || position <= 0 {
		if err != nil {
			return 0, fmt.Errorf("failed to calculate house position: %s", err)
		}
		return 0, fmt.Errorf("failed to calculate house position")
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
