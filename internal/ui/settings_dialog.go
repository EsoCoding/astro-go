package ui

import (
	"astro-go/internal/astro"
	"astro-go/internal/storage"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func showSettingsDialog(win fyne.Window, store *storage.ChartStore, onSaved func()) {
	settings, _ := store.GetSettings()

	dateFormatSelect := widget.NewSelect([]string{"YYYY-MM-DD", "DD-MM-YYYY", "MM-DD-YYYY"}, nil)
	dateFormatSelect.SetSelected(settings.DateFormat)

	timeFormatSelect := widget.NewSelect([]string{"24h", "12h"}, nil)
	timeFormatSelect.SetSelected(settings.TimeFormat)

	locationEntry := widget.NewEntry()
	locationEntry.SetText(settings.DefaultLocation)

	latEntry := widget.NewEntry()
	latEntry.SetText(settings.DefaultLat)

	lngEntry := widget.NewEntry()
	lngEntry.SetText(settings.DefaultLng)

	nodeSelect := widget.NewSelect([]string{"True", "Mean"}, nil)
	nodeSelect.SetSelected(settings.NodePreference)

	pofSelect := widget.NewSelect([]string{"Day", "Day/Night"}, nil)
	pofSelect.SetSelected(settings.PoFPreference)

	enabled := astro.EnabledChartObjectSet(settings.EnabledChartObjects)
	objectChecks := map[astro.Planet]*widget.Check{}
	categoryRows := map[astro.ChartObjectCategory][]fyne.CanvasObject{}
	for _, spec := range astro.ChartObjectCatalog() {
		check := widget.NewCheck(spec.Name, nil)
		check.SetChecked(enabled[spec.Planet])
		objectChecks[spec.Planet] = check
		categoryRows[spec.Category] = append(categoryRows[spec.Category], check)
	}

	generalForm := widget.NewForm(
		widget.NewFormItem("Date Format", dateFormatSelect),
		widget.NewFormItem("Time Format", timeFormatSelect),
		widget.NewFormItem("Default Location", locationEntry),
		widget.NewFormItem("Default Latitude", latEntry),
		widget.NewFormItem("Default Longitude", lngEntry),
		widget.NewFormItem("Node Calculation", nodeSelect),
		widget.NewFormItem("Part of Fortune", pofSelect),
	)
	objectTabs := container.NewAppTabs()
	for _, category := range []astro.ChartObjectCategory{
		astro.ChartObjectCategoryTraditional,
		astro.ChartObjectCategoryModern,
		astro.ChartObjectCategoryNodes,
		astro.ChartObjectCategoryLots,
		astro.ChartObjectCategoryAsteroids,
		astro.ChartObjectCategoryFictitious,
	} {
		rows := categoryRows[category]
		if len(rows) == 0 {
			continue
		}
		objectTabs.Append(container.NewTabItem(string(category), container.NewVScroll(container.NewVBox(rows...))))
	}
	content := container.NewAppTabs(
		container.NewTabItem("General", container.NewVScroll(generalForm)),
		container.NewTabItem("Objects", objectTabs),
	)
	content.SetTabLocation(container.TabLocationTop)

	dlg := dialog.NewCustomConfirm("Settings", "Save", "Cancel", content, func(save bool) {
		if save {
			settings.DateFormat = dateFormatSelect.Selected
			settings.TimeFormat = timeFormatSelect.Selected
			settings.DefaultLocation = locationEntry.Text
			settings.DefaultLat = latEntry.Text
			settings.DefaultLng = lngEntry.Text
			settings.NodePreference = nodeSelect.Selected
			settings.PoFPreference = pofSelect.Selected
			settings.EnabledChartObjects = []astro.Planet{}
			for _, spec := range astro.ChartObjectCatalog() {
				if check, ok := objectChecks[spec.Planet]; ok && check.Checked {
					settings.EnabledChartObjects = append(settings.EnabledChartObjects, spec.Planet)
				}
			}

			store.SaveSettings(settings)
			if onSaved != nil {
				onSaved()
			}
		}
	}, win)

	dlg.Resize(fyne.NewSize(480, 620))
	dlg.Show()
}
