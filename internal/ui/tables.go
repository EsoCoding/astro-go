package ui

import (
	"fmt"

	"astro-go/internal/astro"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func chartSummary(chart astro.Chart) fyne.CanvasObject {
	planetRows := []fyne.CanvasObject{
		widget.NewLabelWithStyle("Planets", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	}
	for _, position := range chart.Planets {
		planetRows = append(planetRows, widget.NewLabel(fmt.Sprintf(
			"%s  %s%s  H%d  %s",
			position.Planet,
			formatZodiac(position.Longitude),
			retrogradeMarker(position.Retrograde),
			position.WholeSignHouse,
			position.EssentialStatus,
		)))
	}

	houseRows := []fyne.CanvasObject{
		widget.NewLabelWithStyle("Whole Sign Houses", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	}
	for _, house := range chart.Houses {
		houseRows = append(houseRows, widget.NewLabel(fmt.Sprintf(
			"H%d  %s  ruler %s",
			house.Number,
			house.Sign,
			house.Ruler,
		)))
	}

	aspectRows := []fyne.CanvasObject{
		widget.NewLabelWithStyle("Traditional Aspects", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	}
	if len(chart.Aspects) == 0 {
		aspectRows = append(aspectRows, widget.NewLabel("None within configured orbs"))
	}
	for _, aspect := range chart.Aspects {
		aspectRows = append(aspectRows, widget.NewLabel(fmt.Sprintf(
			"%s %s %s  orb %.2f",
			aspect.From,
			aspect.Type,
			aspect.To,
			aspect.Orb,
		)))
	}

	return container.NewVScroll(container.NewVBox(
		widget.NewLabelWithStyle(chart.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(chart.DateTimeUTC.Format("2006-01-02 15:04 UTC")),
		widget.NewSeparator(),
		container.NewVBox(planetRows...),
		widget.NewSeparator(),
		container.NewVBox(houseRows...),
		widget.NewSeparator(),
		container.NewVBox(aspectRows...),
	))
}
