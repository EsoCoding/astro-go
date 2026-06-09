package app

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"astro-go/internal/astro"
	"astro-go/internal/storage"
	"astro-go/internal/timezone"
)

type chartCalculator interface {
	NatalChart(data astro.BirthData) (astro.Chart, error)
}

type ChartResolver struct {
	calculator chartCalculator
}

type ResolvedChart struct {
	Single   *astro.Chart
	Synastry *astro.SynastryChart
}

func NewChartResolver(calculator chartCalculator) ChartResolver {
	return ChartResolver{calculator: calculator}
}

func (r ChartResolver) Resolve(saved storage.SavedChart, charts []storage.SavedChart) (ResolvedChart, error) {
	chartType := astro.ChartType(saved.ChartType)
	if chartType == "" {
		chartType = astro.ChartTypeNatal
	}

	switch chartType {
	case astro.ChartTypeNatal:
		data, err := birthDataFromSaved(saved)
		if err != nil {
			return ResolvedChart{}, err
		}
		chart, err := r.calculator.NatalChart(data)
		if err != nil {
			return ResolvedChart{}, err
		}
		return ResolvedChart{Single: &chart}, nil

	case astro.ChartTypeTransit:
		base, ok := findSavedChart(charts, saved.BaseChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("base chart not found for transit definition")
		}
		natalData, err := birthDataFromSaved(base)
		if err != nil {
			return ResolvedChart{}, err
		}
		referenceTime, err := referenceTimeFromSaved(saved)
		if err != nil {
			return ResolvedChart{}, err
		}
		innerChart, err := r.calculator.NatalChart(natalData)
		if err != nil {
			return ResolvedChart{}, err
		}

		transitData := natalData
		transitData.Name = "Transits"
		transitData.DateTimeUTC = referenceTime
		transitData.ChartType = astro.ChartTypeTransit
		if transitData.TimezoneName != "" {
			loc, err := time.LoadLocation(transitData.TimezoneName)
			if err == nil {
				localRef := referenceTime.In(loc)
				_, offsetSec := localRef.Zone()
				offsetHours := float64(offsetSec) / 3600.0
				transitData.UTCOffset = fmt.Sprintf("%g", offsetHours)
			}
		}
		outerChart, err := r.calculator.NatalChart(transitData)
		if err != nil {
			return ResolvedChart{}, err
		}

		synastry := astro.SynastryChart{
			Name:         saved.Name,
			InnerChart:   innerChart,
			OuterChart:   outerChart,
			InterAspects: astro.TraditionalInterAspects(innerChart.Planets, outerChart.Planets),
		}
		return ResolvedChart{Synastry: &synastry}, nil

	case astro.ChartTypeSecondaryProgression:
		base, ok := findSavedChart(charts, saved.BaseChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("base chart not found for progression definition")
		}
		natalData, err := birthDataFromSaved(base)
		if err != nil {
			return ResolvedChart{}, err
		}
		referenceTime, err := referenceTimeFromSaved(saved)
		if err != nil {
			return ResolvedChart{}, err
		}
		innerChart, err := r.calculator.NatalChart(natalData)
		if err != nil {
			return ResolvedChart{}, err
		}

		natalTime := natalData.DateTimeUTC
		duration := referenceTime.Sub(natalTime)
		progressedDuration := time.Duration(float64(duration) / 365.242199)
		progressedTime := natalTime.Add(progressedDuration)

		progressedData := natalData
		progressedData.Name = "Progressed"
		progressedData.DateTimeUTC = progressedTime
		progressedData.ChartType = astro.ChartTypeSecondaryProgression
		outerChart, err := r.calculator.NatalChart(progressedData)
		if err != nil {
			return ResolvedChart{}, err
		}

		synastry := astro.SynastryChart{
			Name:         saved.Name,
			InnerChart:   innerChart,
			OuterChart:   outerChart,
			InterAspects: astro.TraditionalInterAspects(innerChart.Planets, outerChart.Planets),
		}
		return ResolvedChart{Synastry: &synastry}, nil

	case astro.ChartTypeSolarArc:
		base, ok := findSavedChart(charts, saved.BaseChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("base chart not found for solar arc definition")
		}
		natalData, err := birthDataFromSaved(base)
		if err != nil {
			return ResolvedChart{}, err
		}
		referenceTime, err := referenceTimeFromSaved(saved)
		if err != nil {
			return ResolvedChart{}, err
		}
		innerChart, err := r.calculator.NatalChart(natalData)
		if err != nil {
			return ResolvedChart{}, err
		}

		natalTime := natalData.DateTimeUTC
		duration := referenceTime.Sub(natalTime)
		progressedDuration := time.Duration(float64(duration) / 365.242199)
		progressedTime := natalTime.Add(progressedDuration)

		progressedData := natalData
		progressedData.DateTimeUTC = progressedTime
		progressedChart, err := r.calculator.NatalChart(progressedData)
		if err != nil {
			return ResolvedChart{}, err
		}

		var natalSun, progressedSun float64
		for _, p := range innerChart.Planets {
			if p.Planet == astro.Sun {
				natalSun = p.Longitude
			}
		}
		for _, p := range progressedChart.Planets {
			if p.Planet == astro.Sun {
				progressedSun = p.Longitude
			}
		}

		arc := astro.NormalizeDegrees(progressedSun - natalSun)

		outerChart := innerChart
		outerChart.ChartType = astro.ChartTypeSolarArc
		outerChart.Planets = make([]astro.PlanetPosition, len(innerChart.Planets))
		for i, p := range innerChart.Planets {
			shiftedLong := astro.NormalizeDegrees(p.Longitude + arc)
			outerChart.Planets[i] = astro.PlanetPosition{
				Planet:          p.Planet,
				Longitude:       shiftedLong,
				Latitude:        p.Latitude,
				Speed:           p.Speed,
				Sign:            astro.SignFromLongitude(shiftedLong),
				DegreeInSign:    astro.DegreeInSign(shiftedLong),
				House:           p.House,
				Retrograde:      p.Retrograde,
				DomicileRuler:   astro.DomicileRuler(astro.SignFromLongitude(shiftedLong)),
				EssentialStatus: astro.EssentialStatus(p.Planet, astro.SignFromLongitude(shiftedLong)),
			}
		}

		synastry := astro.SynastryChart{
			Name:         saved.Name,
			InnerChart:   innerChart,
			OuterChart:   outerChart,
			InterAspects: astro.TraditionalInterAspects(innerChart.Planets, outerChart.Planets),
		}
		return ResolvedChart{Synastry: &synastry}, nil

	case astro.ChartTypeSolarReturn:
		base, ok := findSavedChart(charts, saved.BaseChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("base chart not found for solar return definition")
		}
		natalData, err := birthDataFromSaved(base)
		if err != nil {
			return ResolvedChart{}, err
		}
		referenceTime, err := referenceTimeFromSaved(saved)
		if err != nil {
			return ResolvedChart{}, err
		}
		innerChart, err := r.calculator.NatalChart(natalData)
		if err != nil {
			return ResolvedChart{}, err
		}

		outerChart, err := r.calculateSolarReturn(natalData, referenceTime)
		if err != nil {
			return ResolvedChart{}, err
		}

		synastry := astro.SynastryChart{
			Name:         saved.Name,
			InnerChart:   innerChart,
			OuterChart:   outerChart,
			InterAspects: astro.TraditionalInterAspects(innerChart.Planets, outerChart.Planets),
		}
		return ResolvedChart{Synastry: &synastry}, nil

	case astro.ChartTypeLunarReturn:
		base, ok := findSavedChart(charts, saved.BaseChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("base chart not found for lunar return definition")
		}
		natalData, err := birthDataFromSaved(base)
		if err != nil {
			return ResolvedChart{}, err
		}
		referenceTime, err := referenceTimeFromSaved(saved)
		if err != nil {
			return ResolvedChart{}, err
		}
		innerChart, err := r.calculator.NatalChart(natalData)
		if err != nil {
			return ResolvedChart{}, err
		}

		outerChart, err := r.calculateLunarReturn(natalData, referenceTime)
		if err != nil {
			return ResolvedChart{}, err
		}

		synastry := astro.SynastryChart{
			Name:         saved.Name,
			InnerChart:   innerChart,
			OuterChart:   outerChart,
			InterAspects: astro.TraditionalInterAspects(innerChart.Planets, outerChart.Planets),
		}
		return ResolvedChart{Synastry: &synastry}, nil

	case astro.ChartTypeSynastry:
		innerSaved, ok := findSavedChart(charts, saved.BaseChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("base chart not found for synastry definition")
		}
		outerSaved, ok := findSavedChart(charts, saved.ComparisonChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("comparison chart not found for synastry definition")
		}
		innerData, err := birthDataFromSaved(innerSaved)
		if err != nil {
			return ResolvedChart{}, err
		}
		outerData, err := birthDataFromSaved(outerSaved)
		if err != nil {
			return ResolvedChart{}, err
		}
		innerData.Name = innerSaved.Name
		outerData.Name = outerSaved.Name
		innerChart, err := r.calculator.NatalChart(innerData)
		if err != nil {
			return ResolvedChart{}, err
		}
		outerChart, err := r.calculator.NatalChart(outerData)
		if err != nil {
			return ResolvedChart{}, err
		}
		synastry := astro.SynastryChart{
			Name:         saved.Name,
			InnerChart:   innerChart,
			OuterChart:   outerChart,
			InterAspects: astro.TraditionalInterAspects(innerChart.Planets, outerChart.Planets),
		}
		return ResolvedChart{Synastry: &synastry}, nil

	default:
		return ResolvedChart{}, fmt.Errorf("%s calculation is not wired yet", chartType.String())
	}
}

func birthDataFromSaved(saved storage.SavedChart) (astro.BirthData, error) {
	localTime, err := time.Parse("2006-01-02 15:04", saved.LocalDate+" "+saved.LocalTime)
	if err != nil {
		return astro.BirthData{}, fmt.Errorf("saved chart %q has invalid date/time", saved.Name)
	}

	offsetHours, err := strconv.ParseFloat(saved.UTCOffset, 64)
	if err != nil {
		return astro.BirthData{}, fmt.Errorf("saved chart %q has invalid UTC offset", saved.Name)
	}
	latitude, err := strconv.ParseFloat(saved.LatitudeDegrees, 64)
	if err != nil {
		return astro.BirthData{}, fmt.Errorf("saved chart %q has invalid latitude", saved.Name)
	}
	longitude, err := strconv.ParseFloat(saved.LongitudeDegrees, 64)
	if err != nil {
		return astro.BirthData{}, fmt.Errorf("saved chart %q has invalid longitude", saved.Name)
	}

	offsetSeconds := int(offsetHours * 3600)
	location := time.FixedZone("chart", offsetSeconds)
	local := time.Date(
		localTime.Year(),
		localTime.Month(),
		localTime.Day(),
		localTime.Hour(),
		localTime.Minute(),
		0,
		0,
		location,
	)

	tzName := timezone.LookupTimezone(latitude, longitude)
	cType := astro.ChartType(saved.ChartType)
	if cType == "" {
		cType = astro.ChartTypeNatal
	}

	return astro.BirthData{
		Name:             saved.Name,
		DateTimeUTC:      local.UTC(),
		LocationName:     saved.LocationName,
		LatitudeDegrees:  latitude,
		LongitudeDegrees: longitude,
		HouseSystem:      astro.HouseSystemFromCode(saved.HouseSystem),
		UTCOffset:        saved.UTCOffset,
		TimezoneName:     tzName,
		ChartType:        cType,
	}, nil
}

func referenceTimeFromSaved(saved storage.SavedChart) (time.Time, error) {
	if saved.ReferenceUTC != "" {
		reference, err := time.Parse(time.RFC3339, saved.ReferenceUTC)
		if err == nil {
			return reference.UTC(), nil
		}
	}
	if saved.ReferenceDate == "" || saved.ReferenceTime == "" {
		return time.Time{}, fmt.Errorf("saved chart %q has no reference time", saved.Name)
	}
	reference, err := time.Parse("2006-01-02 15:04", saved.ReferenceDate+" "+saved.ReferenceTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("saved chart %q has invalid reference date/time", saved.Name)
	}
	return time.Date(reference.Year(), reference.Month(), reference.Day(), reference.Hour(), reference.Minute(), 0, 0, time.UTC), nil
}

func findSavedChart(charts []storage.SavedChart, id string) (storage.SavedChart, bool) {
	for _, chart := range charts {
		if chart.ID == id {
			return chart, true
		}
	}
	return storage.SavedChart{}, false
}

func (r ChartResolver) calculateSolarReturn(natalData astro.BirthData, referenceTime time.Time) (astro.Chart, error) {
	natalChart, err := r.calculator.NatalChart(natalData)
	if err != nil {
		return astro.Chart{}, err
	}
	var natalSun float64
	for _, p := range natalChart.Planets {
		if p.Planet == astro.Sun {
			natalSun = p.Longitude
			break
		}
	}

	natalBirthTime := natalData.DateTimeUTC
	targetYear := referenceTime.Year()
	guessTime := time.Date(targetYear, natalBirthTime.Month(), natalBirthTime.Day(), natalBirthTime.Hour(), natalBirthTime.Minute(), 0, 0, time.UTC)

	currentTime := guessTime
	for iter := 0; iter < 10; iter++ {
		tempData := natalData
		tempData.DateTimeUTC = currentTime
		tempChart, err := r.calculator.NatalChart(tempData)
		if err != nil {
			return astro.Chart{}, err
		}
		var currentSun float64
		for _, p := range tempChart.Planets {
			if p.Planet == astro.Sun {
				currentSun = p.Longitude
				break
			}
		}

		diff := astro.NormalizeDegrees(currentSun - natalSun)
		if diff > 180 {
			diff -= 360
		}
		if math.Abs(diff) < 0.00001 {
			break
		}

		deltaHours := diff / 0.0410686
		currentTime = currentTime.Add(time.Duration(deltaHours * float64(time.Hour)))
	}

	return r.calculator.NatalChart(astro.BirthData{
		Name:             "Solar Return",
		DateTimeUTC:      currentTime,
		LocationName:     natalData.LocationName,
		LatitudeDegrees:  natalData.LatitudeDegrees,
		LongitudeDegrees: natalData.LongitudeDegrees,
		HouseSystem:      natalData.HouseSystem,
		UTCOffset:        natalData.UTCOffset,
		TimezoneName:     natalData.TimezoneName,
		ChartType:        astro.ChartTypeSolarReturn,
	})
}

func (r ChartResolver) calculateLunarReturn(natalData astro.BirthData, referenceTime time.Time) (astro.Chart, error) {
	natalChart, err := r.calculator.NatalChart(natalData)
	if err != nil {
		return astro.Chart{}, err
	}
	var natalMoon float64
	for _, p := range natalChart.Planets {
		if p.Planet == astro.Moon {
			natalMoon = p.Longitude
			break
		}
	}

	currentTime := referenceTime
	for iter := 0; iter < 15; iter++ {
		tempData := natalData
		tempData.DateTimeUTC = currentTime
		tempChart, err := r.calculator.NatalChart(tempData)
		if err != nil {
			return astro.Chart{}, err
		}
		var currentMoon float64
		for _, p := range tempChart.Planets {
			if p.Planet == astro.Moon {
				currentMoon = p.Longitude
				break
			}
		}

		diff := astro.NormalizeDegrees(currentMoon - natalMoon)
		if diff > 180 {
			diff -= 360
		}
		if math.Abs(diff) < 0.00001 {
			break
		}

		deltaHours := diff / 0.549
		currentTime = currentTime.Add(time.Duration(deltaHours * float64(time.Hour)))
	}

	return r.calculator.NatalChart(astro.BirthData{
		Name:             "Lunar Return",
		DateTimeUTC:      currentTime,
		LocationName:     natalData.LocationName,
		LatitudeDegrees:  natalData.LatitudeDegrees,
		LongitudeDegrees: natalData.LongitudeDegrees,
		HouseSystem:      natalData.HouseSystem,
		UTCOffset:        natalData.UTCOffset,
		TimezoneName:     natalData.TimezoneName,
		ChartType:        astro.ChartTypeLunarReturn,
	})
}
