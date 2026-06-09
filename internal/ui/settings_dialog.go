package ui

import (
	"astro-go/internal/storage"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func showSettingsDialog(win fyne.Window, store *storage.ChartStore) {
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

	form := widget.NewForm(
		widget.NewFormItem("Date Format", dateFormatSelect),
		widget.NewFormItem("Time Format", timeFormatSelect),
		widget.NewFormItem("Default Location", locationEntry),
		widget.NewFormItem("Default Latitude", latEntry),
		widget.NewFormItem("Default Longitude", lngEntry),
		widget.NewFormItem("Node Calculation", nodeSelect),
		widget.NewFormItem("Part of Fortune", pofSelect),
	)

	dlg := dialog.NewCustomConfirm("Settings", "Save", "Cancel", container.NewVScroll(form), func(save bool) {
		if save {
			settings.DateFormat = dateFormatSelect.Selected
			settings.TimeFormat = timeFormatSelect.Selected
			settings.DefaultLocation = locationEntry.Text
			settings.DefaultLat = latEntry.Text
			settings.DefaultLng = lngEntry.Text
			settings.NodePreference = nodeSelect.Selected
			settings.PoFPreference = pofSelect.Selected

			store.SaveSettings(settings)
		}
	}, win)

	dlg.Resize(fyne.NewSize(400, 350))
	dlg.Show()
}
