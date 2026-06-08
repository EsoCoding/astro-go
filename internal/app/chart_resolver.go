package app

import (
	"fmt"
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
		baseData, err := birthDataFromSaved(base)
		if err != nil {
			return ResolvedChart{}, err
		}
		referenceTime, err := referenceTimeFromSaved(saved)
		if err != nil {
			return ResolvedChart{}, err
		}
		baseData.Name = saved.Name
		baseData.DateTimeUTC = referenceTime
		baseData.ChartType = astro.ChartTypeTransit
		if baseData.TimezoneName != "" {
			loc, err := time.LoadLocation(baseData.TimezoneName)
			if err == nil {
				localRef := referenceTime.In(loc)
				_, offsetSec := localRef.Zone()
				offsetHours := float64(offsetSec) / 3600.0
				baseData.UTCOffset = fmt.Sprintf("%g", offsetHours)
			}
		}
		chart, err := r.calculator.NatalChart(baseData)
		if err != nil {
			return ResolvedChart{}, err
		}
		return ResolvedChart{Single: &chart}, nil
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
