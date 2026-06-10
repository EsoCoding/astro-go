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
	calculator     chartCalculator
	enabledObjects []astro.Planet
}

type ResolvedChart struct {
	Single   *astro.Chart
	Synastry *astro.SynastryChart
}

func NewChartResolver(calculator chartCalculator) ChartResolver {
	return ChartResolver{calculator: calculator}
}

func (r *ChartResolver) SetEnabledObjects(objects []astro.Planet) {
	r.enabledObjects = append([]astro.Planet(nil), objects...)
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
		data.EnabledObjects = r.enabledObjects
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
		natalData.EnabledObjects = r.enabledObjects
		referenceTime, err := referenceTimeFromSaved(saved)
		if err != nil {
			return ResolvedChart{}, err
		}

		transitData := natalData
		transitData.Name = saved.Name
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
		return ResolvedChart{Single: &outerChart}, nil

	case astro.ChartTypeSecondaryProgression:
		base, ok := findSavedChart(charts, saved.BaseChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("base chart not found for progression definition")
		}
		natalData, err := birthDataFromSaved(base)
		if err != nil {
			return ResolvedChart{}, err
		}
		natalData.EnabledObjects = r.enabledObjects
		referenceTime, err := referenceTimeFromSaved(saved)
		if err != nil {
			return ResolvedChart{}, err
		}

		natalTime := natalData.DateTimeUTC
		duration := referenceTime.Sub(natalTime)
		progressedDuration := time.Duration(float64(duration) / 365.242199)
		progressedTime := natalTime.Add(progressedDuration)

		progressedData := natalData
		progressedData.Name = saved.Name
		progressedData.DateTimeUTC = progressedTime
		progressedData.ChartType = astro.ChartTypeSecondaryProgression
		outerChart, err := r.calculator.NatalChart(progressedData)
		if err != nil {
			return ResolvedChart{}, err
		}
		return ResolvedChart{Single: &outerChart}, nil

	case astro.ChartTypeSolarArc:
		base, ok := findSavedChart(charts, saved.BaseChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("base chart not found for solar arc definition")
		}
		natalData, err := birthDataFromSaved(base)
		if err != nil {
			return ResolvedChart{}, err
		}
		natalData.EnabledObjects = r.enabledObjects
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

		arcData := natalData
		arcData.EnabledObjects = withRequiredObjects(arcData.EnabledObjects, astro.Sun)
		arcInnerChart, err := r.calculator.NatalChart(arcData)
		if err != nil {
			return ResolvedChart{}, err
		}
		progressedData := arcData
		progressedData.DateTimeUTC = progressedTime
		progressedChart, err := r.calculator.NatalChart(progressedData)
		if err != nil {
			return ResolvedChart{}, err
		}

		var natalSun, progressedSun float64
		for _, p := range arcInnerChart.Planets {
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
		outerChart.Name = saved.Name
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
		outerChart.Aspects = astro.TraditionalAspects(outerChart.Planets)
		return ResolvedChart{Single: &outerChart}, nil

	case astro.ChartTypeSolarReturn:
		base, ok := findSavedChart(charts, saved.BaseChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("base chart not found for solar return definition")
		}
		natalData, err := birthDataFromSaved(base)
		if err != nil {
			return ResolvedChart{}, err
		}
		natalData.EnabledObjects = r.enabledObjects

		// Relocation override
		if saved.RelocatedLatitude != "" && saved.RelocatedLongitude != "" {
			if lat, err := strconv.ParseFloat(saved.RelocatedLatitude, 64); err == nil {
				if lng, err := strconv.ParseFloat(saved.RelocatedLongitude, 64); err == nil {
					natalData.LatitudeDegrees = lat
					natalData.LongitudeDegrees = lng
					natalData.TimezoneName = timezone.LookupTimezone(lat, lng)
					if saved.RelocatedLocationName != "" {
						natalData.LocationName = saved.RelocatedLocationName
					}
				}
			}
		}
		referenceTime, err := referenceTimeFromSaved(saved)
		if err != nil {
			return ResolvedChart{}, err
		}

		outerChart, err := r.calculateSolarReturn(natalData, referenceTime, saved.Name)
		if err != nil {
			return ResolvedChart{}, err
		}
		return ResolvedChart{Single: &outerChart}, nil

	case astro.ChartTypeLunarReturn:
		base, ok := findSavedChart(charts, saved.BaseChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("base chart not found for lunar return definition")
		}
		natalData, err := birthDataFromSaved(base)
		if err != nil {
			return ResolvedChart{}, err
		}
		natalData.EnabledObjects = r.enabledObjects

		// Relocation override
		if saved.RelocatedLatitude != "" && saved.RelocatedLongitude != "" {
			if lat, err := strconv.ParseFloat(saved.RelocatedLatitude, 64); err == nil {
				if lng, err := strconv.ParseFloat(saved.RelocatedLongitude, 64); err == nil {
					natalData.LatitudeDegrees = lat
					natalData.LongitudeDegrees = lng
					natalData.TimezoneName = timezone.LookupTimezone(lat, lng)
					if saved.RelocatedLocationName != "" {
						natalData.LocationName = saved.RelocatedLocationName
					}
				}
			}
		}
		referenceTime, err := referenceTimeFromSaved(saved)
		if err != nil {
			return ResolvedChart{}, err
		}

		outerChart, err := r.calculateLunarReturn(natalData, referenceTime, saved.Name)
		if err != nil {
			return ResolvedChart{}, err
		}
		return ResolvedChart{Single: &outerChart}, nil

	case astro.ChartTypeSynastry:
		innerSaved, ok := findSavedChart(charts, saved.BaseChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("base chart not found for synastry definition")
		}
		outerSaved, ok := findSavedChart(charts, saved.ComparisonChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("comparison chart not found for synastry definition")
		}
		resolvedInner, err := r.Resolve(innerSaved, charts)
		if err != nil {
			return ResolvedChart{}, err
		}
		resolvedOuter, err := r.Resolve(outerSaved, charts)
		if err != nil {
			return ResolvedChart{}, err
		}
		if resolvedInner.Single == nil {
			return ResolvedChart{}, fmt.Errorf("inner chart of synastry must be a single chart")
		}
		if resolvedOuter.Single == nil {
			return ResolvedChart{}, fmt.Errorf("outer chart of synastry must be a single chart")
		}
		synastry := astro.SynastryChart{
			Name:         saved.Name,
			InnerChart:   *resolvedInner.Single,
			OuterChart:   *resolvedOuter.Single,
			InterAspects: astro.TraditionalInterAspects(resolvedInner.Single.Planets, resolvedOuter.Single.Planets),
		}
		return ResolvedChart{Synastry: &synastry}, nil

	case astro.ChartTypeTertiaryProgression:
		base, ok := findSavedChart(charts, saved.BaseChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("base chart not found for progression definition")
		}
		natalData, err := birthDataFromSaved(base)
		if err != nil {
			return ResolvedChart{}, err
		}
		natalData.EnabledObjects = r.enabledObjects
		referenceTime, err := referenceTimeFromSaved(saved)
		if err != nil {
			return ResolvedChart{}, err
		}

		natalTime := natalData.DateTimeUTC
		duration := referenceTime.Sub(natalTime)
		// Rate: 1 day = 1 sidereal month (27.321661 days)
		progressedDuration := time.Duration(float64(duration) / 27.321661)
		progressedTime := natalTime.Add(progressedDuration)

		progressedData := natalData
		progressedData.Name = saved.Name
		progressedData.DateTimeUTC = progressedTime
		progressedData.ChartType = astro.ChartTypeTertiaryProgression
		outerChart, err := r.calculator.NatalChart(progressedData)
		if err != nil {
			return ResolvedChart{}, err
		}
		return ResolvedChart{Single: &outerChart}, nil

	case astro.ChartTypePrimaryDirection:
		base, ok := findSavedChart(charts, saved.BaseChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("base chart not found for primary direction definition")
		}
		natalData, err := birthDataFromSaved(base)
		if err != nil {
			return ResolvedChart{}, err
		}
		natalData.EnabledObjects = r.enabledObjects
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
		years := float64(duration) / (365.242199 * 24 * float64(time.Hour))

		rate := 0.9855556 // Naibod key: 59'08" per year
		if saved.DirectionKey == "ptolemy" {
			rate = 1.0 // Ptolemy key: 1 degree per year
		}
		arc := astro.NormalizeDegrees(years * rate)

		outerChart := innerChart
		outerChart.ChartType = astro.ChartTypePrimaryDirection
		outerChart.Name = saved.Name
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
		outerChart.Aspects = astro.TraditionalAspects(outerChart.Planets)
		return ResolvedChart{Single: &outerChart}, nil

	case astro.ChartTypeComposite:
		innerSaved, ok := findSavedChart(charts, saved.BaseChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("base chart not found for composite definition")
		}
		outerSaved, ok := findSavedChart(charts, saved.ComparisonChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("comparison chart not found for composite definition")
		}
		resolvedInner, err := r.Resolve(innerSaved, charts)
		if err != nil {
			return ResolvedChart{}, err
		}
		resolvedOuter, err := r.Resolve(outerSaved, charts)
		if err != nil {
			return ResolvedChart{}, err
		}
		if resolvedInner.Single == nil {
			return ResolvedChart{}, fmt.Errorf("inner chart of composite must be a single chart")
		}
		if resolvedOuter.Single == nil {
			return ResolvedChart{}, fmt.Errorf("outer chart of composite must be a single chart")
		}
		innerChart := *resolvedInner.Single
		outerChart := *resolvedOuter.Single

		compositeChart := innerChart
		compositeChart.Name = saved.Name
		compositeChart.ChartType = astro.ChartTypeComposite

		compositeChart.Planets = make([]astro.PlanetPosition, len(innerChart.Planets))
		for i, ip := range innerChart.Planets {
			var op astro.PlanetPosition
			found := false
			for _, p := range outerChart.Planets {
				if p.Planet == ip.Planet {
					op = p
					found = true
					break
				}
			}
			if !found {
				compositeChart.Planets[i] = ip
				continue
			}
			midLong := angleMidpoint(ip.Longitude, op.Longitude)
			compositeChart.Planets[i] = astro.PlanetPosition{
				Planet:        ip.Planet,
				Longitude:     midLong,
				Latitude:      (ip.Latitude + op.Latitude) / 2,
				Speed:         (ip.Speed + op.Speed) / 2,
				Sign:          astro.SignFromLongitude(midLong),
				DegreeInSign:  astro.DegreeInSign(midLong),
				House:         ip.House,
				Retrograde:    ip.Retrograde || op.Retrograde,
				DomicileRuler: astro.DomicileRuler(astro.SignFromLongitude(midLong)),
			}
		}

		compositeChart.Houses = make([]astro.House, len(innerChart.Houses))
		for i, ih := range innerChart.Houses {
			var oh astro.House
			found := false
			for _, h := range outerChart.Houses {
				if h.Number == ih.Number {
					oh = h
					found = true
					break
				}
			}
			if !found {
				compositeChart.Houses[i] = ih
				continue
			}
			midLong := angleMidpoint(ih.CuspLongitude, oh.CuspLongitude)
			compositeChart.Houses[i] = astro.House{
				Number:        ih.Number,
				CuspLongitude: midLong,
				Sign:          astro.SignFromLongitude(midLong),
				Ruler:         astro.DomicileRuler(astro.SignFromLongitude(midLong)),
			}
		}

		ascLong := angleMidpoint(innerChart.Ascendant.Longitude, outerChart.Ascendant.Longitude)
		compositeChart.Ascendant = astro.Angle{
			Name:         innerChart.Ascendant.Name,
			Longitude:    ascLong,
			Sign:         astro.SignFromLongitude(ascLong),
			DegreeInSign: astro.DegreeInSign(ascLong),
		}
		mcLong := angleMidpoint(innerChart.MC.Longitude, outerChart.MC.Longitude)
		compositeChart.MC = astro.Angle{
			Name:         innerChart.MC.Name,
			Longitude:    mcLong,
			Sign:         astro.SignFromLongitude(mcLong),
			DegreeInSign: astro.DegreeInSign(mcLong),
		}
		compositeChart.Aspects = astro.TraditionalAspects(compositeChart.Planets)

		return ResolvedChart{Single: &compositeChart}, nil

	case astro.ChartTypeDavison:
		innerSaved, ok := findSavedChart(charts, saved.BaseChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("base chart not found for Davison definition")
		}
		outerSaved, ok := findSavedChart(charts, saved.ComparisonChartID)
		if !ok {
			return ResolvedChart{}, fmt.Errorf("comparison chart not found for Davison definition")
		}
		innerData, err := birthDataFromSaved(innerSaved)
		if err != nil {
			return ResolvedChart{}, err
		}
		outerData, err := birthDataFromSaved(outerSaved)
		if err != nil {
			return ResolvedChart{}, err
		}

		t1 := innerData.DateTimeUTC
		t2 := outerData.DateTimeUTC
		midTime := t1.Add(t2.Sub(t1) / 2)

		midLat := (innerData.LatitudeDegrees + outerData.LatitudeDegrees) / 2
		midLng := longitudeMidpoint(innerData.LongitudeDegrees, outerData.LongitudeDegrees)

		davisonData := innerData
		davisonData.Name = saved.Name
		davisonData.DateTimeUTC = midTime
		davisonData.LatitudeDegrees = midLat
		davisonData.LongitudeDegrees = midLng
		davisonData.LocationName = "Relationship Midpoint"
		davisonData.ChartType = astro.ChartTypeDavison
		davisonData.EnabledObjects = r.enabledObjects

		chart, err := r.calculator.NatalChart(davisonData)
		if err != nil {
			return ResolvedChart{}, err
		}
		return ResolvedChart{Single: &chart}, nil

	default:
		return ResolvedChart{}, fmt.Errorf("%s calculation is not wired yet", chartType.String())
	}
}

func angleMidpoint(a, b float64) float64 {
	diff := math.Abs(a - b)
	if diff <= 180 {
		return astro.NormalizeDegrees((a + b) / 2)
	}
	return astro.NormalizeDegrees((a + b + 360) / 2)
}

func longitudeMidpoint(lng1, lng2 float64) float64 {
	l1 := lng1 + 180
	l2 := lng2 + 180
	mid := angleMidpoint(l1, l2)
	return mid - 180
}

func birthDataFromSaved(saved storage.SavedChart) (astro.BirthData, error) {
	localTime, err := parseSavedDateTime(saved.LocalDate, saved.LocalTime)
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
		localTime.Second(),
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
	reference, err := parseSavedDateTime(saved.ReferenceDate, saved.ReferenceTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("saved chart %q has invalid reference date/time", saved.Name)
	}
	return time.Date(reference.Year(), reference.Month(), reference.Day(), reference.Hour(), reference.Minute(), reference.Second(), 0, time.UTC), nil
}

func parseSavedDateTime(date, clock string) (time.Time, error) {
	if parsed, err := time.Parse("2006-01-02 15:04:05", date+" "+clock); err == nil {
		return parsed, nil
	}
	return time.Parse("2006-01-02 15:04", date+" "+clock)
}

func findSavedChart(charts []storage.SavedChart, id string) (storage.SavedChart, bool) {
	for _, chart := range charts {
		if chart.ID == id {
			return chart, true
		}
	}
	return storage.SavedChart{}, false
}

func withRequiredObjects(objects []astro.Planet, required ...astro.Planet) []astro.Planet {
	next := append([]astro.Planet(nil), objects...)
	seen := astro.EnabledChartObjectSet(next)
	for _, object := range required {
		if !seen[object] {
			next = append(next, object)
			seen[object] = true
		}
	}
	return next
}

func (r ChartResolver) calculateSolarReturn(natalData astro.BirthData, referenceTime time.Time, name string) (astro.Chart, error) {
	referenceData := natalData
	referenceData.EnabledObjects = withRequiredObjects(referenceData.EnabledObjects, astro.Sun)
	natalChart, err := r.calculator.NatalChart(referenceData)
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
		tempData.EnabledObjects = referenceData.EnabledObjects
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
		Name:             name,
		DateTimeUTC:      currentTime,
		LocationName:     natalData.LocationName,
		LatitudeDegrees:  natalData.LatitudeDegrees,
		LongitudeDegrees: natalData.LongitudeDegrees,
		HouseSystem:      natalData.HouseSystem,
		UTCOffset:        natalData.UTCOffset,
		TimezoneName:     natalData.TimezoneName,
		ChartType:        astro.ChartTypeSolarReturn,
		EnabledObjects:   natalData.EnabledObjects,
	})
}

func (r ChartResolver) calculateLunarReturn(natalData astro.BirthData, referenceTime time.Time, name string) (astro.Chart, error) {
	referenceData := natalData
	referenceData.EnabledObjects = withRequiredObjects(referenceData.EnabledObjects, astro.Moon)
	natalChart, err := r.calculator.NatalChart(referenceData)
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
		tempData.EnabledObjects = referenceData.EnabledObjects
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
		Name:             name,
		DateTimeUTC:      currentTime,
		LocationName:     natalData.LocationName,
		LatitudeDegrees:  natalData.LatitudeDegrees,
		LongitudeDegrees: natalData.LongitudeDegrees,
		HouseSystem:      natalData.HouseSystem,
		UTCOffset:        natalData.UTCOffset,
		TimezoneName:     natalData.TimezoneName,
		ChartType:        astro.ChartTypeLunarReturn,
		EnabledObjects:   natalData.EnabledObjects,
	})
}
