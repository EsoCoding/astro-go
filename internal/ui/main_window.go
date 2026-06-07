package ui

import (
	"fmt"
	"strconv"
	"time"

	"astro-go/internal/astro"
	"astro-go/internal/sweph"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func Launch() {
	application := app.New()
	window := application.NewWindow("Astro Go")
	window.Resize(fyne.NewSize(1180, 760))

	calculator := sweph.NewCalculator()
	initial := astro.BirthData{
		Name:             "Example Natal Chart",
		DateTimeUTC:      time.Date(1990, 1, 1, 12, 0, 0, 0, time.UTC),
		LatitudeDegrees:  52.3676,
		LongitudeDegrees: 4.9041,
	}
	chart, err := calculator.NatalChart(initial)
	if err != nil {
		window.SetContent(widget.NewLabel(err.Error()))
		window.ShowAndRun()
		return
	}

	name := widget.NewEntry()
	name.SetText(initial.Name)
	date := widget.NewEntry()
	date.SetText("1990-01-01")
	clock := widget.NewEntry()
	clock.SetText("12:00")
	timezone := widget.NewEntry()
	timezone.SetText("0")
	latitude := widget.NewEntry()
	latitude.SetText("52.3676")
	longitude := widget.NewEntry()
	longitude.SetText("4.9041")

	wheelSlot := container.NewCenter(NewChartWheel(chart))
	summarySlot := container.NewStack(chartSummary(chart))
	status := widget.NewLabel("")

	calculate := widget.NewButton("Calculate", func() {
		data, parseErr := parseBirthData(name.Text, date.Text, clock.Text, timezone.Text, latitude.Text, longitude.Text)
		if parseErr != nil {
			status.SetText(parseErr.Error())
			return
		}
		nextChart, calcErr := calculator.NatalChart(data)
		if calcErr != nil {
			status.SetText(calcErr.Error())
			return
		}
		wheelSlot.Objects = []fyne.CanvasObject{NewChartWheel(nextChart)}
		wheelSlot.Refresh()
		summarySlot.Objects = []fyne.CanvasObject{chartSummary(nextChart)}
		summarySlot.Refresh()
		status.SetText("")
	})

	form := widget.NewForm(
		widget.NewFormItem("Name", name),
		widget.NewFormItem("Date", date),
		widget.NewFormItem("Time", clock),
		widget.NewFormItem("UTC offset", timezone),
		widget.NewFormItem("Latitude", latitude),
		widget.NewFormItem("Longitude", longitude),
	)
	left := container.NewBorder(nil, container.NewVBox(calculate, status), nil, nil, form)
	left.Resize(fyne.NewSize(300, 760))

	content := container.NewBorder(
		nil,
		nil,
		left,
		summarySlot,
		wheelSlot,
	)
	window.SetContent(content)
	window.ShowAndRun()
}

func parseBirthData(name, date, clock, timezoneOffset, latitudeValue, longitudeValue string) (astro.BirthData, error) {
	localTime, err := time.Parse("2006-01-02 15:04", date+" "+clock)
	if err != nil {
		return astro.BirthData{}, fmt.Errorf("date/time must use YYYY-MM-DD and HH:MM")
	}

	offsetHours, err := strconv.ParseFloat(timezoneOffset, 64)
	if err != nil {
		return astro.BirthData{}, fmt.Errorf("UTC offset must be a number")
	}
	latitude, err := strconv.ParseFloat(latitudeValue, 64)
	if err != nil {
		return astro.BirthData{}, fmt.Errorf("latitude must be a number")
	}
	longitude, err := strconv.ParseFloat(longitudeValue, 64)
	if err != nil {
		return astro.BirthData{}, fmt.Errorf("longitude must be a number")
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

	return astro.BirthData{
		Name:             name,
		DateTimeUTC:      local.UTC(),
		LatitudeDegrees:  latitude,
		LongitudeDegrees: longitude,
	}, nil
}
