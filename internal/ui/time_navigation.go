package ui

import (
	"fmt"
	"time"

	"astro-go/internal/astro"
	"astro-go/internal/storage"
)

const (
	timeStepUnitSecond = "Second"
	timeStepUnitMinute = "Minute"
	timeStepUnitHour   = "Hour"
	timeStepUnitDay    = "Day"
	timeStepUnitWeek   = "Week"
	timeStepUnitMonth  = "Month"
	timeStepUnitYear   = "Year"
)

func timeStepUnits() []string {
	return []string{
		timeStepUnitSecond,
		timeStepUnitMinute,
		timeStepUnitHour,
		timeStepUnitDay,
		timeStepUnitWeek,
		timeStepUnitMonth,
		timeStepUnitYear,
	}
}

func stepTime(value time.Time, unit string, amount int) time.Time {
	switch unit {
	case timeStepUnitSecond:
		return value.Add(time.Duration(amount) * time.Second)
	case timeStepUnitMinute:
		return value.Add(time.Duration(amount) * time.Minute)
	case timeStepUnitHour:
		return value.Add(time.Duration(amount) * time.Hour)
	case timeStepUnitDay:
		return value.AddDate(0, 0, amount)
	case timeStepUnitWeek:
		return value.AddDate(0, 0, amount*7)
	case timeStepUnitMonth:
		return value.AddDate(0, amount, 0)
	case timeStepUnitYear:
		return value.AddDate(amount, 0, 0)
	default:
		return value
	}
}

func workspaceTitle(chartType astro.ChartType) string {
	if chartType == "" {
		chartType = astro.ChartTypeNatal
	}
	return fmt.Sprintf("%s Workspace", chartType.String())
}

func chartTypeFromSaved(chart storage.SavedChart) astro.ChartType {
	if chart.ChartType == "" {
		return astro.ChartTypeNatal
	}
	return astro.ChartType(chart.ChartType)
}

func activeTimeText(chartType astro.ChartType, chart storage.SavedChart, hasChart bool, date, clock string) string {
	if chartType.RequiresReferenceTime() && hasChart {
		referenceTime, err := referenceTimeFromChart(chart)
		if err == nil {
			return fmt.Sprintf("Reference %s UTC", formatDateTime(referenceTime))
		}
	}
	if chartType == astro.ChartTypeNatal {
		return fmt.Sprintf("Local %s %s", date, clock)
	}
	return ""
}

func referenceTimeFromChart(chart storage.SavedChart) (time.Time, error) {
	if chart.ReferenceUTC != "" {
		referenceTime, err := time.Parse(time.RFC3339, chart.ReferenceUTC)
		if err == nil {
			return referenceTime.UTC(), nil
		}
	}
	return parseReferenceDateTime(chart.ReferenceDate, chart.ReferenceTime)
}

func parseLocalDateTime(date, clock string) (time.Time, error) {
	if parsed, err := time.Parse("2006-01-02 15:04:05", date+" "+clock); err == nil {
		return parsed, nil
	}
	return time.Parse("2006-01-02 15:04", date+" "+clock)
}

func formatClock(value time.Time) string {
	if value.Second() != 0 {
		return value.Format("15:04:05")
	}
	return value.Format("15:04")
}

func formatDateTime(value time.Time) string {
	return fmt.Sprintf("%s %s", value.Format("2006-01-02"), formatClock(value))
}

func swapSynastry(chart astro.SynastryChart) astro.SynastryChart {
	chart.InnerChart, chart.OuterChart = chart.OuterChart, chart.InnerChart
	chart.InterAspects = astro.TraditionalInterAspects(chart.InnerChart.Planets, chart.OuterChart.Planets)
	return chart
}
