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
	planets, err := planetPositions(julianDay, data.LatitudeDegrees, asc, armc, eps, houseSystem, data.EnabledObjects)
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

func planetPositions(julianDay, latitude, asc, armc, eps float64, houseSystem astro.HouseSystem, enabledObjects []astro.Planet) ([]astro.PlanetPosition, error) {
	flags := int32(swisseph.FlagSwieph | swisseph.FlagSpeed)
	enabled := astro.EnabledChartObjectSet(enabledObjects)
	positions := make([]astro.PlanetPosition, 0, len(enabled))
	calculated := map[astro.Planet]astro.PlanetPosition{}

	for _, spec := range astro.ChartObjectCatalog() {
		planet := spec.Planet
		if !enabled[planet] || isManualPlanet(planet) {
			continue
		}
		position, ok, err := calculatePlanetPosition(julianDay, latitude, armc, eps, houseSystem, planet, flags)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		calculated[planet] = position
		positions = append(positions, position)
	}

	if enabled[astro.SouthNode] {
		nodePlanet := astro.NorthNode
		if enabled[astro.MeanSouthNode] && !enabled[astro.NorthNode] {
			nodePlanet = astro.MeanNorthNode
		}
		northNode, err := calculatedOrHiddenPosition(calculated, julianDay, latitude, armc, eps, houseSystem, nodePlanet, flags)
		if err != nil {
			return nil, err
		}
		position := calculatedSouthNode(northNode, astro.SouthNode, latitude, armc, eps, houseSystem)
		calculated[astro.SouthNode] = position
		positions = append(positions, position)
	}

	if enabled[astro.MeanSouthNode] {
		northNode, err := calculatedOrHiddenPosition(calculated, julianDay, latitude, armc, eps, houseSystem, astro.MeanNorthNode, flags)
		if err != nil {
			return nil, err
		}
		position := calculatedSouthNode(northNode, astro.MeanSouthNode, latitude, armc, eps, houseSystem)
		calculated[astro.MeanSouthNode] = position
		positions = append(positions, position)
	}

	if enabled[astro.ParsFortunae] {
		sun, err := calculatedOrHiddenPosition(calculated, julianDay, latitude, armc, eps, houseSystem, astro.Sun, flags)
		if err != nil {
			return nil, err
		}
		moon, err := calculatedOrHiddenPosition(calculated, julianDay, latitude, armc, eps, houseSystem, astro.Moon, flags)
		if err != nil {
			return nil, err
		}
		position := calculatedParsFortunae(sun.Longitude, moon.Longitude, asc, latitude, armc, eps, houseSystem)
		calculated[astro.ParsFortunae] = position
		positions = append(positions, position)
	}

	return positions, nil
}

func calculatePlanetPosition(julianDay, latitude, armc, eps float64, houseSystem astro.HouseSystem, planet astro.Planet, flags int32) (astro.PlanetPosition, bool, error) {
	swephID, ok := swephIDForPlanet(planet)
	if !ok {
		return astro.PlanetPosition{}, false, fmt.Errorf("no Swiss Ephemeris mapping for %s", planet)
	}
	result := swisseph.CalcUT(julianDay, swephID, flags)
	if result.Flag < 0 {
		if isRequiredPlanet(planet) {
			return astro.PlanetPosition{}, false, fmt.Errorf("failed to calculate %s: %s", planet, result.Error)
		}
		return astro.PlanetPosition{}, false, nil
	}
	values := result.Data
	longitude := astro.NormalizeDegrees(values[0])
	position, err := positionFromValues(planet, longitude, values[1], values[3], latitude, armc, eps, houseSystem)
	if err != nil {
		if !isRequiredPlanet(planet) {
			return astro.PlanetPosition{}, false, nil
		}
		return astro.PlanetPosition{}, false, err
	}
	return position, true, nil
}

func calculatedOrHiddenPosition(calculated map[astro.Planet]astro.PlanetPosition, julianDay, latitude, armc, eps float64, houseSystem astro.HouseSystem, planet astro.Planet, flags int32) (astro.PlanetPosition, error) {
	if position, ok := calculated[planet]; ok {
		return position, nil
	}
	position, ok, err := calculatePlanetPosition(julianDay, latitude, armc, eps, houseSystem, planet, flags)
	if err != nil {
		return astro.PlanetPosition{}, err
	}
	if !ok {
		return astro.PlanetPosition{}, fmt.Errorf("failed to calculate required dependency %s", planet)
	}
	calculated[planet] = position
	return position, nil
}

func calculatedSouthNode(northNode astro.PlanetPosition, planet astro.Planet, latitude, armc, eps float64, houseSystem astro.HouseSystem) astro.PlanetPosition {
	longitude := astro.NormalizeDegrees(northNode.Longitude + 180.0)
	position, _ := positionFromValues(planet, longitude, 0, northNode.Speed, latitude, armc, eps, houseSystem)
	position.Retrograde = true
	return position
}

func calculatedParsFortunae(sunLong, moonLong, asc, latitude, armc, eps float64, houseSystem astro.HouseSystem) astro.PlanetPosition {
	isDayChart := astro.NormalizeDegrees(sunLong-asc) >= 180
	longitude := astro.NormalizeDegrees(asc + moonLong - sunLong)
	if !isDayChart {
		longitude = astro.NormalizeDegrees(asc + sunLong - moonLong)
	}
	position, _ := positionFromValues(astro.ParsFortunae, longitude, 0, 0, latitude, armc, eps, houseSystem)
	return position
}

func positionFromValues(planet astro.Planet, longitude, bodyLatitude, speed, geolat, armc, eps float64, houseSystem astro.HouseSystem) (astro.PlanetPosition, error) {
	sign := astro.SignFromLongitude(longitude)
	house, err := planetHouse(longitude, bodyLatitude, geolat, armc, eps, houseSystem)
	if err != nil {
		return astro.PlanetPosition{}, err
	}
	essentialStatus := ""
	if spec, ok := astro.ChartObjectSpecFor(planet); ok && spec.HasDignity {
		essentialStatus = astro.EssentialStatus(planet, sign)
	}
	return astro.PlanetPosition{
		Planet:          planet,
		Longitude:       longitude,
		Latitude:        bodyLatitude,
		Speed:           speed,
		Sign:            sign,
		DegreeInSign:    astro.DegreeInSign(longitude),
		House:           house,
		Retrograde:      speed < 0,
		DomicileRuler:   astro.DomicileRuler(sign),
		EssentialStatus: essentialStatus,
	}, nil
}

func isManualPlanet(planet astro.Planet) bool {
	return planet == astro.SouthNode || planet == astro.MeanSouthNode || planet == astro.ParsFortunae
}

func isRequiredPlanet(planet astro.Planet) bool {
	for _, traditional := range astro.TraditionalPlanets {
		if planet == traditional {
			return true
		}
	}
	return false
}

func swephIDForPlanet(planet astro.Planet) (int32, bool) {
	swephIDs := map[astro.Planet]int32{
		astro.Sun:                 swisseph.Sun,
		astro.Moon:                swisseph.Moon,
		astro.Mercury:             swisseph.Mercury,
		astro.Venus:               swisseph.Venus,
		astro.Mars:                swisseph.Mars,
		astro.Jupiter:             swisseph.Jupiter,
		astro.Saturn:              swisseph.Saturn,
		astro.Uranus:              swisseph.Uranus,
		astro.Neptune:             swisseph.Neptune,
		astro.Pluto:               swisseph.Pluto,
		astro.NorthNode:           swisseph.TrueNode,
		astro.MeanNorthNode:       swisseph.MeanNode,
		astro.BlackMoonLilith:     swisseph.MeanApog,
		astro.TrueBlackMoonLilith: swisseph.OscuApog,
		astro.Earth:               swisseph.Earth,
		astro.Chiron:              swisseph.Chiron,
		astro.Pholus:              swisseph.Pholus,
		astro.Ceres:               swisseph.Ceres,
		astro.Pallas:              swisseph.Pallas,
		astro.Juno:                swisseph.Juno,
		astro.Vesta:               swisseph.Vesta,
		astro.Varuna:              swisseph.Varuna,
		astro.Cupido:              swisseph.Cupido,
		astro.Hades:               swisseph.Hades,
		astro.Zeus:                swisseph.Zeus,
		astro.Kronos:              swisseph.Kronos,
		astro.Apollon:             swisseph.Apollon,
		astro.Admetos:             swisseph.Admetos,
		astro.Vulkanus:            swisseph.Vulkanus,
		astro.Poseidon:            swisseph.Poseidon,
		astro.Isis:                swisseph.Isis,
		astro.WhiteMoon:           swisseph.WhiteMoon,
		astro.Proserpina:          swisseph.Proserpina,
	}
	id, ok := swephIDs[planet]
	return id, ok
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
