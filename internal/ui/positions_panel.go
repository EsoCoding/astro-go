package ui

import (
	"fmt"

	"astro-go/internal/astro"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// buildNatalPositions creates a side panel for a standard natal chart
func buildNatalPositions(chart astro.Chart) fyne.CanvasObject {
	planetTable := widget.NewTable(
		func() (int, int) { return len(chart.Planets) + 1, 5 },
		func() fyne.CanvasObject { return widget.NewLabel("Template Text Long") },
		func(i widget.TableCellID, o fyne.CanvasObject) {
			l := o.(*widget.Label)
			l.TextStyle = fyne.TextStyle{Monospace: true}
			if i.Row == 0 {
				l.TextStyle.Bold = true
				switch i.Col {
				case 0:
					l.SetText("Planet")
				case 1:
					l.SetText("Position")
				case 2:
					l.SetText("House")
				case 3:
					l.SetText("Dignity")
				case 4:
					l.SetText("Retro")
				}
				return
			}
			l.TextStyle.Bold = false
			p := chart.Planets[i.Row-1]
			switch i.Col {
			case 0:
				l.SetText(shortPlanetName(p.Planet))
			case 1:
				l.SetText(formatZodiacDMS(p.Longitude))
			case 2:
				l.SetText(fmt.Sprintf("H%d", p.House))
			case 3:
				l.SetText(string(p.EssentialStatus))
			case 4:
				l.SetText(retrogradeMarker(p.Retrograde))
			}
		},
	)
	planetTable.SetColumnWidth(0, 80)
	planetTable.SetColumnWidth(1, 130)
	planetTable.SetColumnWidth(2, 60)
	planetTable.SetColumnWidth(3, 80)
	planetTable.SetColumnWidth(4, 50)

	houseTable := widget.NewTable(
		func() (int, int) { return len(chart.Houses) + 1, 3 },
		func() fyne.CanvasObject { return widget.NewLabel("Template Text Long") },
		func(i widget.TableCellID, o fyne.CanvasObject) {
			l := o.(*widget.Label)
			l.TextStyle = fyne.TextStyle{Monospace: true}
			if i.Row == 0 {
				l.TextStyle.Bold = true
				switch i.Col {
				case 0:
					l.SetText("House")
				case 1:
					l.SetText("Cusp")
				case 2:
					l.SetText("Ruler")
				}
				return
			}
			l.TextStyle.Bold = false
			h := chart.Houses[i.Row-1]
			switch i.Col {
			case 0:
				l.SetText(fmt.Sprintf("%d", h.Number))
			case 1:
				l.SetText(formatZodiacDMS(h.CuspLongitude))
			case 2:
				l.SetText(shortPlanetName(h.Ruler))
			}
		},
	)
	houseTable.SetColumnWidth(0, 60)
	houseTable.SetColumnWidth(1, 130)
	houseTable.SetColumnWidth(2, 80)

	tabs := container.NewAppTabs(
		container.NewTabItem("Planets", planetTable),
		container.NewTabItem("Houses", houseTable),
	)
	return tabs
}

func buildSynastryPositions(chart astro.SynastryChart) fyne.CanvasObject {
	aspectTable := widget.NewTable(
		func() (int, int) { return len(chart.InterAspects) + 1, 3 },
		func() fyne.CanvasObject { return widget.NewLabel("Template Text Long") },
		func(i widget.TableCellID, o fyne.CanvasObject) {
			l := o.(*widget.Label)
			l.TextStyle = fyne.TextStyle{Monospace: true}
			if i.Row == 0 {
				l.TextStyle.Bold = true
				switch i.Col {
				case 0:
					l.SetText("Inner")
				case 1:
					l.SetText("Aspect")
				case 2:
					l.SetText("Outer")
				}
				return
			}
			l.TextStyle.Bold = false
			a := chart.InterAspects[i.Row-1]
			switch i.Col {
			case 0:
				l.SetText(shortPlanetName(a.Inner))
			case 1:
				l.SetText(aspectGlyph(a.Type))
			case 2:
				l.SetText(shortPlanetName(a.Outer))
			}
		},
	)
	aspectTable.SetColumnWidth(0, 100)
	aspectTable.SetColumnWidth(1, 80)
	aspectTable.SetColumnWidth(2, 100)

	tabs := container.NewAppTabs(
		container.NewTabItem("Inter Aspects", aspectTable),
	)
	return tabs
}
