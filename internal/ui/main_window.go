package ui

import (
	"fmt"
	"strconv"
	"time"

	astroapp "astro-go/internal/app"
	"astro-go/internal/astro"
	"astro-go/internal/geocode"
	"astro-go/internal/storage"
	"astro-go/internal/sweph"
	"astro-go/internal/timezone"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func Launch() {
	application := fyneapp.NewWithID("com.esocode.astro-go")
	window := application.NewWindow("Astro Go")
	window.Resize(fyne.NewSize(1280, 820))

	calculator := sweph.NewCalculator()
	geocoder := geocode.NewNominatimClient()
	resolver := astroapp.NewChartResolver(calculator)
	store, err := storage.NewChartStore(application.Preferences())
	if err != nil {
		window.SetContent(widget.NewLabel(err.Error()))
		window.ShowAndRun()
		return
	}
	defer store.Close()
	activeSettings, _ := store.GetSettings()
	resolver.SetEnabledObjects(activeSettings.EnabledChartObjects)
	enabledChartObjects := func() []astro.Planet {
		settings, _ := store.GetSettings()
		activeSettings = settings
		resolver.SetEnabledObjects(settings.EnabledChartObjects)
		return settings.EnabledChartObjects
	}
	savedCharts, err := store.List()
	if err != nil {
		savedCharts = nil
	}
	selectedSavedID := ""
	var initial astro.BirthData
	var initialDateStr, initialTimeStr, initialOffsetStr string

	if len(savedCharts) > 0 {
		firstSaved := savedCharts[0]
		selectedSavedID = firstSaved.ID
		initialDateStr = firstSaved.LocalDate
		initialTimeStr = firstSaved.LocalTime
		initialOffsetStr = firstSaved.UTCOffset

		resolved, err := resolver.Resolve(firstSaved, savedCharts)
		if err == nil {
			if resolved.Single != nil {
				initial = astro.BirthData{
					Name:             firstSaved.Name,
					DateTimeUTC:      resolved.Single.DateTimeUTC,
					LocationName:     resolved.Single.LocationName,
					LatitudeDegrees:  resolved.Single.Latitude,
					LongitudeDegrees: resolved.Single.Longitude,
					HouseSystem:      resolved.Single.HouseSystem,
					EnabledObjects:   activeSettings.EnabledChartObjects,
				}
			} else if resolved.Synastry != nil {
				initial = astro.BirthData{
					Name:             firstSaved.Name,
					DateTimeUTC:      resolved.Synastry.InnerChart.DateTimeUTC,
					LocationName:     resolved.Synastry.InnerChart.LocationName,
					LatitudeDegrees:  resolved.Synastry.InnerChart.Latitude,
					LongitudeDegrees: resolved.Synastry.InnerChart.Longitude,
					HouseSystem:      resolved.Synastry.InnerChart.HouseSystem,
					EnabledObjects:   activeSettings.EnabledChartObjects,
				}
			}
		}
	}

	if initial.Name == "" {
		settings, _ := store.GetSettings()

		loc, err := time.LoadLocation("Europe/Amsterdam")
		if err != nil {
			loc = time.UTC
		}
		nowLocal := time.Now().In(loc)
		_, offsetSec := nowLocal.Zone()
		offsetHours := float64(offsetSec) / 3600.0

		initialDateStr = nowLocal.Format("2006-01-02")
		initialTimeStr = nowLocal.Format("15:04")
		initialOffsetStr = fmt.Sprintf("%g", offsetHours)

		defaultLat := 52.3676
		defaultLng := 4.9041
		defaultLocName := "Amsterdam, Netherlands"

		if settings.DefaultLat != "" && settings.DefaultLng != "" {
			if lat, err := strconv.ParseFloat(settings.DefaultLat, 64); err == nil {
				defaultLat = lat
			}
			if lng, err := strconv.ParseFloat(settings.DefaultLng, 64); err == nil {
				defaultLng = lng
			}
			if settings.DefaultLocation != "" {
				defaultLocName = settings.DefaultLocation
			} else {
				defaultLocName = "Custom Location"
			}
		}

		initial = astro.BirthData{
			Name:             "Example Natal Chart",
			DateTimeUTC:      nowLocal.UTC(),
			LocationName:     defaultLocName,
			LatitudeDegrees:  defaultLat,
			LongitudeDegrees: defaultLng,
			HouseSystem:      astro.DefaultHouseSystem(),
			EnabledObjects:   activeSettings.EnabledChartObjects,
		}
	}
	initial.EnabledObjects = activeSettings.EnabledChartObjects

	chart, err := calculator.NatalChart(initial)
	if err != nil {
		window.SetContent(widget.NewLabel(err.Error()))
		window.ShowAndRun()
		return
	}
	currentChart := chart

	name := widget.NewEntry()
	name.SetText(initial.Name)
	name.SetPlaceHolder("Chart name")
	date := widget.NewEntry()
	date.SetText(initialDateStr)
	date.SetPlaceHolder("YYYY-MM-DD")
	clock := widget.NewEntry()
	clock.SetText(initialTimeStr)
	clock.SetPlaceHolder("HH:MM")
	timezone := widget.NewEntry()
	timezone.SetText(initialOffsetStr)
	timezone.SetPlaceHolder("UTC offset, e.g. 1 or -5")
	locationName := widget.NewEntry()
	locationName.SetText(initial.LocationName)
	locationName.SetPlaceHolder("Place name")
	latitude := widget.NewEntry()
	latitude.SetText(fmt.Sprintf("%.6f", initial.LatitudeDegrees))
	latitude.SetPlaceHolder("Decimal degrees")
	longitude := widget.NewEntry()
	longitude.SetText(fmt.Sprintf("%.6f", initial.LongitudeDegrees))
	longitude.SetPlaceHolder("Decimal degrees")

	houseSystem := widget.NewSelect(astro.HouseSystemOptions(), nil)
	houseSystem.SetSelected(initial.HouseSystem.Label())
	zodiacMode := widget.NewSelect([]string{"Tropical"}, nil)
	zodiacMode.SetSelected("Tropical")

	wheelSlot := container.NewStack(NewChartWheel(chart))
	positionsSlot := container.NewStack(buildNatalPositions(chart))
	chartArea := container.NewHSplit(container.NewPadded(wheelSlot), container.NewPadded(positionsSlot))
	chartArea.Offset = 0.75
	chartBody := chartArea
	status := widget.NewLabel("Ready")
	currentInputSummary := widget.NewLabel(currentChartInputSummary(name.Text, date.Text, clock.Text))
	var chartList *widget.List
	activeChartType := astro.ChartTypeNatal
	activeSavedChart := storage.SavedChart{}
	hasActiveSavedChart := false
	var activeSynastryChart *astro.SynastryChart
	synastrySwapped := false
	var workspaceLabel *widget.Label
	var activeTimeLabel *widget.Label
	var stepAmount *widget.Entry
	var stepUnit *widget.Select
	var backButton *widget.Button
	var forwardButton *widget.Button
	var swapButton *widget.Button
	var editButton *widget.Button
	var loadSavedChart func(storage.SavedChart)

	updateWorkspaceChrome := func() {
		if workspaceLabel != nil {
			workspaceLabel.SetText(workspaceTitle(activeChartType))
		}
		if activeTimeLabel != nil {
			activeTimeLabel.SetText(activeTimeText(activeChartType, activeSavedChart, hasActiveSavedChart, date.Text, clock.Text))
		}
		canNavigate := activeChartType == astro.ChartTypeNatal || activeChartType.RequiresReferenceTime()
		if backButton != nil {
			if canNavigate {
				backButton.Enable()
			} else {
				backButton.Disable()
			}
		}
		if forwardButton != nil {
			if canNavigate {
				forwardButton.Enable()
			} else {
				forwardButton.Disable()
			}
		}
		if swapButton != nil {
			if activeSynastryChart != nil {
				swapButton.Show()
				swapButton.Enable()
			} else {
				swapButton.Hide()
			}
		}
		if editButton != nil {
			if hasActiveSavedChart && activeChartType != astro.ChartTypeSynastry {
				editButton.Enable()
			} else {
				editButton.Disable()
			}
		}
	}

	setFormFromSavedChart := func(saved storage.SavedChart) {
		selectedSavedID = saved.ID
		name.SetText(saved.Name)
		date.SetText(saved.LocalDate)
		clock.SetText(saved.LocalTime)
		timezone.SetText(saved.UTCOffset)
		locationName.SetText(shortenLocationName(saved.LocationName))
		latitude.SetText(saved.LatitudeDegrees)
		longitude.SetText(saved.LongitudeDegrees)
		houseSystem.SetSelected(astro.HouseSystemFromCode(saved.HouseSystem).Label())
		currentInputSummary.SetText(currentChartInputSummary(name.Text, date.Text, clock.Text))
	}

	refreshSavedCharts := func() {
		charts, loadErr := store.List()
		if loadErr != nil {
			status.SetText(loadErr.Error())
			return
		}
		savedCharts = charts
		if selectedSavedID != "" {
			if _, ok := savedChartByID(savedCharts, selectedSavedID); !ok {
				selectedSavedID = ""
			}
		}
		if chartList != nil {
			chartList.Refresh()
		}
	}

	saveChart := func(updateSelected bool) {
		if updateSelected && selectedSavedID != "" {
			if selected, ok := savedChartByID(savedCharts, selectedSavedID); ok && !canCalculateSavedChart(selected) {
				status.SetText(fmt.Sprintf("%s definitions are not editable from the natal birth form", chartTypeLabel(selected.ChartType)))
				return
			}
		}
		data, parseErr := parseBirthData(name.Text, date.Text, clock.Text, timezone.Text, locationName.Text, latitude.Text, longitude.Text, houseSystem.Selected)
		if parseErr != nil {
			status.SetText(parseErr.Error())
			return
		}
		saved := storage.SavedChartFromBirthData(data, date.Text, clock.Text, timezone.Text, locationName.Text, latitude.Text, longitude.Text)
		if updateSelected && selectedSavedID != "" {
			saved.ID = selectedSavedID
		}
		if saveErr := store.Save(&saved); saveErr != nil {
			status.SetText(saveErr.Error())
			return
		}
		selectedSavedID = saved.ID
		refreshSavedCharts()
		if chartList != nil {
			if selectedIndex := savedChartIndexByID(savedCharts, selectedSavedID); selectedIndex >= 0 {
				chartList.Select(selectedIndex)
			}
		}
		if selected, ok := savedChartByID(savedCharts, selectedSavedID); ok {
			loadSavedChart(selected)
		}
		status.SetText(fmt.Sprintf("Saved chart %s", saved.Name))
	}

	saveCurrentChartAsNew := func() {
		saveChart(false)
	}

	deleteSavedChart := func() {
		if selectedSavedID == "" {
			status.SetText("Select a saved chart first")
			return
		}
		if deleteErr := store.Delete(selectedSavedID); deleteErr != nil {
			status.SetText(deleteErr.Error())
			return
		}
		selectedSavedID = ""
		refreshSavedCharts()
		if chartList != nil {
			chartList.UnselectAll()
		}
		status.SetText("Deleted saved chart")
	}

	refreshChart := func(c astro.Chart) {
		currentChart = c
		activeSynastryChart = nil
		wheelSlot.Objects = []fyne.CanvasObject{NewChartWheel(c)}
		wheelSlot.Refresh()
		positionsSlot.Objects = []fyne.CanvasObject{buildNatalPositions(c)}
		positionsSlot.Refresh()
		updateWorkspaceChrome()
	}

	refreshSynastryChart := func(synastry astro.SynastryChart) {
		currentChart = synastry.InnerChart
		activeSynastryChart = &synastry
		wheelSlot.Objects = []fyne.CanvasObject{NewSynastryWheel(synastry)}
		wheelSlot.Refresh()
		positionsSlot.Objects = []fyne.CanvasObject{buildSynastryPositions(synastry)}
		positionsSlot.Refresh()
		updateWorkspaceChrome()
		status.SetText(fmt.Sprintf("Loaded %s %s x %s", activeChartType.String(), synastry.InnerChart.Name, synastry.OuterChart.Name))
	}

	loadSavedChart = func(selected storage.SavedChart) {
		selectedSavedID = selected.ID
		activeSavedChart = selected
		hasActiveSavedChart = true
		activeChartType = chartTypeFromSaved(selected)
		synastrySwapped = false
		enabledChartObjects()
		resolvedChart, err := resolver.Resolve(selected, savedCharts)
		if err == nil {
			if resolvedChart.Synastry != nil {
				refreshSynastryChart(*resolvedChart.Synastry)
				return
			}
			if resolvedChart.Single != nil && canCalculateSavedChart(selected) {
				setFormFromSavedChart(selected)
			}
			if resolvedChart.Single != nil {
				refreshChart(*resolvedChart.Single)
				status.SetText(fmt.Sprintf("Loaded %s", selected.Name))
				return
			}
			if canCalculateSavedChart(selected) {
				setFormFromSavedChart(selected)
				status.SetText("No chart data resolved")
				return
			}
		}
		if canCalculateSavedChart(selected) {
			setFormFromSavedChart(selected)
			status.SetText(err.Error())
			updateWorkspaceChrome()
			return
		}
		status.SetText(err.Error())
		updateWorkspaceChrome()
	}

	calculateActiveChart := func() bool {
		data, parseErr := parseBirthData(name.Text, date.Text, clock.Text, timezone.Text, locationName.Text, latitude.Text, longitude.Text, houseSystem.Selected)
		if parseErr != nil {
			status.SetText(parseErr.Error())
			return false
		}
		data.EnabledObjects = enabledChartObjects()
		nextChart, calcErr := calculator.NatalChart(data)
		if calcErr != nil {
			status.SetText(calcErr.Error())
			return false
		}
		currentInputSummary.SetText(currentChartInputSummary(name.Text, date.Text, clock.Text))
		activeChartType = astro.ChartTypeNatal
		hasActiveSavedChart = false
		synastrySwapped = false
		refreshChart(nextChart)
		return true
	}

	navigateTime := func(direction int) {
		amount, err := strconv.Atoi(stepAmount.Text)
		if err != nil || amount <= 0 {
			status.SetText("Step must be a positive whole number")
			return
		}
		unit := stepUnit.Selected
		if unit == "" {
			unit = timeStepUnitMinute
		}
		amount *= direction

		if activeChartType == astro.ChartTypeNatal {
			localTime, err := parseLocalDateTime(date.Text, clock.Text)
			if err != nil {
				status.SetText("date/time must use YYYY-MM-DD and HH:MM or HH:MM:SS")
				return
			}
			nextTime := stepTime(localTime, unit, amount)
			date.SetText(nextTime.Format("2006-01-02"))
			clock.SetText(formatClock(nextTime))
			if calculateActiveChart() {
				saveChart(selectedSavedID != "")
				status.SetText(fmt.Sprintf("Moved natal chart to %s %s", date.Text, clock.Text))
			}
			return
		}

		if !activeChartType.RequiresReferenceTime() {
			status.SetText(fmt.Sprintf("%s has no navigable reference time", activeChartType.String()))
			return
		}
		if !hasActiveSavedChart {
			status.SetText("Select a saved derived chart first")
			return
		}
		referenceTime, err := referenceTimeFromChart(activeSavedChart)
		if err != nil {
			status.SetText(err.Error())
			return
		}
		nextTime := stepTime(referenceTime, unit, amount)
		activeSavedChart.ReferenceDate = nextTime.Format("2006-01-02")
		activeSavedChart.ReferenceTime = formatClock(nextTime)
		activeSavedChart.ReferenceUTC = nextTime.Format(time.RFC3339)
		if saveErr := store.Save(&activeSavedChart); saveErr != nil {
			status.SetText(saveErr.Error())
			return
		}
		refreshSavedCharts()
		enabledChartObjects()
		resolvedChart, err := resolver.Resolve(activeSavedChart, savedCharts)
		if err != nil {
			status.SetText(err.Error())
			updateWorkspaceChrome()
			return
		}
		if resolvedChart.Synastry != nil {
			nextSynastry := *resolvedChart.Synastry
			if synastrySwapped {
				nextSynastry = swapSynastry(nextSynastry)
			}
			refreshSynastryChart(nextSynastry)
		} else if resolvedChart.Single != nil {
			refreshChart(*resolvedChart.Single)
		}
		status.SetText(fmt.Sprintf("Moved %s reference to %s %s UTC", activeChartType.String(), activeSavedChart.ReferenceDate, activeSavedChart.ReferenceTime))
	}

	swapActiveSynastry := func() {
		if activeSynastryChart == nil {
			status.SetText("Select a synastry-style chart first")
			return
		}
		swapped := swapSynastry(*activeSynastryChart)
		synastrySwapped = !synastrySwapped
		refreshSynastryChart(swapped)
		status.SetText(fmt.Sprintf("Swapped Synastry %s x %s", swapped.InnerChart.Name, swapped.OuterChart.Name))
	}

	showNewNatalChartDialog := func() {
		dataWindow := application.NewWindow("New Natal Chart")
		dataWindow.Resize(fyne.NewSize(420, 360))

		nextName := newTabbableEntry()
		nextDate := newTabbableEntry()
		nextTime := newTabbableEntry()
		nextOffset := newTabbableEntry()
		nextLocation := newTabbableEntry()
		nextLatitude := newTabbableEntry()
		nextLongitude := newTabbableEntry()
		nextHouseSystem := widget.NewSelect(astro.HouseSystemOptions(), nil)

		nextName.SetText(defaultChartName(astro.ChartTypeNatal))
		nextDate.SetText(time.Now().Format("2006-01-02"))
		nextTime.SetText("12:00")
		nextOffset.SetText("0")
		settings, _ := store.GetSettings()
		locText := locationName.Text
		latText := latitude.Text
		lngText := longitude.Text

		if settings.DefaultLocation != "" || settings.DefaultLat != "" {
			if settings.DefaultLocation != "" {
				locText = settings.DefaultLocation
			}
			if settings.DefaultLat != "" {
				latText = settings.DefaultLat
				lngText = settings.DefaultLng
			}
		}

		nextLocation.SetText(locText)
		nextLatitude.SetText(latText)
		nextLongitude.SetText(lngText)
		nextHouseSystem.SetSelected(houseSystem.Selected)
		lookupLocation := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
			result, err := geocoder.Lookup(nextLocation.Text)
			if err != nil {
				dialog.ShowError(err, dataWindow)
				return
			}
			nextLocation.SetText(result.DisplayName)
			nextLatitude.SetText(fmt.Sprintf("%.6f", result.Latitude))
			nextLongitude.SetText(fmt.Sprintf("%.6f", result.Longitude))
			status.SetText(fmt.Sprintf("Resolved %s", result.DisplayName))
		})

		onCuspsChanged := func(string) {
			updateOffset(&nextDate.Entry, &nextTime.Entry, &nextLatitude.Entry, &nextLongitude.Entry, &nextOffset.Entry, status)
		}
		nextDate.OnChanged = onCuspsChanged
		nextTime.OnChanged = onCuspsChanged
		nextLatitude.OnChanged = onCuspsChanged
		nextLongitude.OnChanged = onCuspsChanged

		updateOffset(&nextDate.Entry, &nextTime.Entry, &nextLatitude.Entry, &nextLongitude.Entry, &nextOffset.Entry, status)

		create := func() {
			selectedSavedID = ""
			name.SetText(nextName.Text)
			date.SetText(nextDate.Text)
			clock.SetText(nextTime.Text)
			timezone.SetText(nextOffset.Text)
			locationName.SetText(nextLocation.Text)
			latitude.SetText(nextLatitude.Text)
			longitude.SetText(nextLongitude.Text)
			houseSystem.SetSelected(nextHouseSystem.Selected)
			dataWindow.Close()
			if calculateActiveChart() {
				saveCurrentChartAsNew()
			}
		}

		onSubmit := func(string) {
			create()
		}
		nextName.OnSubmitted = onSubmit
		nextDate.OnSubmitted = onSubmit
		nextTime.OnSubmitted = onSubmit
		nextOffset.OnSubmitted = onSubmit
		nextLocation.OnSubmitted = onSubmit
		nextLatitude.OnSubmitted = onSubmit
		nextLongitude.OnSubmitted = onSubmit

		createLabeledField := func(labelText string, input fyne.CanvasObject) fyne.CanvasObject {
			lbl := widget.NewLabelWithStyle(labelText, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			return container.NewVBox(lbl, input)
		}

		nameField := createLabeledField("Name", nextName)
		dateField := createLabeledField("Date", nextDate)
		timeField := createLabeledField("Time", nextTime)
		offsetField := createLabeledField("UTC Offset", nextOffset)
		locationField := createLabeledField("Location", container.NewBorder(nil, nil, nil, lookupLocation, nextLocation))
		latitudeField := createLabeledField("Latitude", nextLatitude)
		longitudeField := createLabeledField("Longitude", nextLongitude)
		housesField := createLabeledField("Houses", nextHouseSystem)

		formLayout := container.NewVBox(
			nameField,
			container.NewGridWithColumns(2, dateField, timeField),
			locationField,
			container.NewGridWithColumns(3, latitudeField, longitudeField, offsetField),
			housesField,
		)

		buttons := container.NewGridWithColumns(2,
			widget.NewButton("Cancel", dataWindow.Close),
			widget.NewButton("Create", create),
		)

		dataWindow.SetContent(widget.NewCard("Natal Chart", "Create a natal chart from birth data.", container.NewBorder(
			nil,
			buttons,
			nil,
			nil,
			formLayout,
		)))
		dataWindow.Show()
	}

	var showBirthDataDialog func()

	getNatalCharts := func() ([]string, map[string]storage.SavedChart) {
		options := []string{}
		lookup := map[string]storage.SavedChart{}
		for _, saved := range savedCharts {
			if saved.ChartType != "" && saved.ChartType != string(astro.ChartTypeNatal) {
				continue
			}
			label := savedChartLabel(saved)
			options = append(options, label)
			lookup[label] = saved
		}
		return options, lookup
	}

	createNatalSelector := func(options []string, lookup map[string]storage.SavedChart) *widget.Select {
		sel := widget.NewSelect(options, nil)
		if len(options) > 0 {
			if hasActiveSavedChart && activeSavedChart.ChartType == string(astro.ChartTypeNatal) {
				activeLabel := savedChartLabel(activeSavedChart)
				for _, opt := range options {
					if opt == activeLabel {
						sel.SetSelected(opt)
						break
					}
				}
			}
			if sel.Selected == "" {
				sel.SetSelected(options[0])
			}
		}
		return sel
	}

	showSolarReturnDialog := func(editChart *storage.SavedChart) {
		isEdit := editChart != nil
		titleText := "Cast Solar Return Chart"
		actionText := "Cast"
		if isEdit {
			titleText = "Edit Solar Return Chart"
			actionText = "Save"
		}

		dataWindow := application.NewWindow(titleText)
		dataWindow.Resize(fyne.NewSize(450, 420))

		options, lookup := getNatalCharts()
		if len(options) == 0 {
			dialog.ShowInformation("No Natal Charts", "Please create a natal chart first.", window)
			return
		}

		baseSelect := createNatalSelector(options, lookup)

		yearEntry := widget.NewEntry()
		yearEntry.SetText(fmt.Sprintf("%d", time.Now().Year()))

		nameEntry := widget.NewEntry()
		updateName := func() {
			if !isEdit {
				nameEntry.SetText(fmt.Sprintf("Solar Return %s %s", lookup[baseSelect.Selected].Name, yearEntry.Text))
			}
		}
		baseSelect.OnChanged = func(value string) { updateName() }
		yearEntry.OnChanged = func(value string) { updateName() }

		relocateCheck := widget.NewCheck("Relocate return chart", nil)
		relocateLocation := widget.NewEntry()
		relocateLocation.SetPlaceHolder("Location name")
		relocateLat := widget.NewEntry()
		relocateLat.SetPlaceHolder("Latitude")
		relocateLon := widget.NewEntry()
		relocateLon.SetPlaceHolder("Longitude")

		statusLabel := widget.NewLabel("")

		relocateSearch := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
			result, err := geocoder.Lookup(relocateLocation.Text)
			if err != nil {
				statusLabel.SetText("Geocoding failed: " + err.Error())
				return
			}
			relocateLocation.SetText(result.DisplayName)
			relocateLat.SetText(fmt.Sprintf("%.6f", result.Latitude))
			relocateLon.SetText(fmt.Sprintf("%.6f", result.Longitude))
		})

		relocateFields := container.NewVBox(
			widget.NewLabelWithStyle("Relocation Coordinates", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewBorder(nil, nil, nil, relocateSearch, relocateLocation),
			container.NewGridWithColumns(2, relocateLat, relocateLon),
		)
		relocateFields.Hide()

		relocateCheck.OnChanged = func(checked bool) {
			if checked {
				relocateFields.Show()
			} else {
				relocateFields.Hide()
			}
		}

		if isEdit {
			nameEntry.SetText(editChart.Name)
			if baseChart, ok := savedChartByID(savedCharts, editChart.BaseChartID); ok {
				baseSelect.SetSelected(savedChartLabel(baseChart))
			}
			if editChart.ReferenceDate != "" {
				if parsed, err := time.Parse("2006-01-02", editChart.ReferenceDate); err == nil {
					yearEntry.SetText(fmt.Sprintf("%d", parsed.Year()))
				}
			}
			if editChart.RelocatedLatitude != "" && editChart.RelocatedLongitude != "" {
				relocateCheck.SetChecked(true)
				relocateLocation.SetText(editChart.RelocatedLocationName)
				relocateLat.SetText(editChart.RelocatedLatitude)
				relocateLon.SetText(editChart.RelocatedLongitude)
				relocateFields.Show()
			}
		} else {
			updateName()
		}

		cast := func() {
			base := lookup[baseSelect.Selected]
			natalTime, err := parseLocalDateTime(base.LocalDate, base.LocalTime)
			if err != nil {
				statusLabel.SetText("Error parsing natal birth time: " + err.Error())
				return
			}

			year, err := strconv.Atoi(yearEntry.Text)
			if err != nil || year < 1000 || year > 9999 {
				statusLabel.SetText("Invalid return year")
				return
			}

			var saved storage.SavedChart
			if isEdit {
				saved = *editChart
			} else {
				saved = base
				saved.ID = ""
			}
			saved.Name = nameEntry.Text
			saved.ChartType = string(astro.ChartTypeSolarReturn)
			saved.BaseChartID = base.ID
			saved.ComparisonChartID = ""
			saved.ReferenceDate = fmt.Sprintf("%04d-%02d-%02d", year, int(natalTime.Month()), natalTime.Day())
			saved.ReferenceTime = formatClock(natalTime)
			saved.ReferenceUTC = time.Date(year, natalTime.Month(), natalTime.Day(), natalTime.Hour(), natalTime.Minute(), natalTime.Second(), 0, time.UTC).Format(time.RFC3339)
			saved.DirectionKey = ""

			saved.RelocatedLatitude = ""
			saved.RelocatedLongitude = ""
			saved.RelocatedLocationName = ""
			if relocateCheck.Checked {
				saved.RelocatedLatitude = relocateLat.Text
				saved.RelocatedLongitude = relocateLon.Text
				saved.RelocatedLocationName = relocateLocation.Text
			}

			if saveErr := store.Save(&saved); saveErr != nil {
				statusLabel.SetText(saveErr.Error())
				return
			}
			dataWindow.Close()
			selectedSavedID = saved.ID
			refreshSavedCharts()
			if chartList != nil {
				if selectedIndex := savedChartIndexByID(savedCharts, selectedSavedID); selectedIndex >= 0 {
					chartList.Select(selectedIndex)
				}
			}
			loadSavedChart(saved)
			status.SetText(fmt.Sprintf("Saved solar return chart %s", saved.Name))
		}

		form := widget.NewForm(
			widget.NewFormItem("Base Natal Chart", baseSelect),
			widget.NewFormItem("Chart Name", nameEntry),
			widget.NewFormItem("Return Year", yearEntry),
		)

		buttons := container.NewGridWithColumns(2,
			widget.NewButton("Cancel", dataWindow.Close),
			widget.NewButton(actionText, cast),
		)

		content := container.NewVBox(form, relocateCheck, relocateFields, statusLabel, buttons)
		dataWindow.SetContent(widget.NewCard(titleText, "Configure solar return parameters.", content))
		dataWindow.Show()
	}

	showLunarReturnDialog := func(editChart *storage.SavedChart) {
		isEdit := editChart != nil
		titleText := "Cast Lunar Return Chart"
		actionText := "Cast"
		if isEdit {
			titleText = "Edit Lunar Return Chart"
			actionText = "Save"
		}

		dataWindow := application.NewWindow(titleText)
		dataWindow.Resize(fyne.NewSize(450, 450))

		options, lookup := getNatalCharts()
		if len(options) == 0 {
			dialog.ShowInformation("No Natal Charts", "Please create a natal chart first.", window)
			return
		}

		baseSelect := createNatalSelector(options, lookup)

		yearEntry := widget.NewEntry()
		yearEntry.SetText(fmt.Sprintf("%d", time.Now().Year()))

		months := []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
		monthSelect := widget.NewSelect(months, nil)
		monthSelect.SetSelected(time.Now().Month().String())

		nameEntry := widget.NewEntry()
		updateName := func() {
			if !isEdit {
				nameEntry.SetText(fmt.Sprintf("Lunar Return %s %s %s", lookup[baseSelect.Selected].Name, yearEntry.Text, monthSelect.Selected))
			}
		}
		baseSelect.OnChanged = func(value string) { updateName() }
		yearEntry.OnChanged = func(value string) { updateName() }
		monthSelect.OnChanged = func(value string) { updateName() }

		relocateCheck := widget.NewCheck("Relocate return chart", nil)
		relocateLocation := widget.NewEntry()
		relocateLocation.SetPlaceHolder("Location name")
		relocateLat := widget.NewEntry()
		relocateLat.SetPlaceHolder("Latitude")
		relocateLon := widget.NewEntry()
		relocateLon.SetPlaceHolder("Longitude")

		statusLabel := widget.NewLabel("")

		relocateSearch := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
			result, err := geocoder.Lookup(relocateLocation.Text)
			if err != nil {
				statusLabel.SetText("Geocoding failed: " + err.Error())
				return
			}
			relocateLocation.SetText(result.DisplayName)
			relocateLat.SetText(fmt.Sprintf("%.6f", result.Latitude))
			relocateLon.SetText(fmt.Sprintf("%.6f", result.Longitude))
		})

		relocateFields := container.NewVBox(
			widget.NewLabelWithStyle("Relocation Coordinates", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewBorder(nil, nil, nil, relocateSearch, relocateLocation),
			container.NewGridWithColumns(2, relocateLat, relocateLon),
		)
		relocateFields.Hide()

		relocateCheck.OnChanged = func(checked bool) {
			if checked {
				relocateFields.Show()
			} else {
				relocateFields.Hide()
			}
		}

		if isEdit {
			nameEntry.SetText(editChart.Name)
			if baseChart, ok := savedChartByID(savedCharts, editChart.BaseChartID); ok {
				baseSelect.SetSelected(savedChartLabel(baseChart))
			}
			if editChart.ReferenceDate != "" {
				if parsed, err := time.Parse("2006-01-02", editChart.ReferenceDate); err == nil {
					yearEntry.SetText(fmt.Sprintf("%d", parsed.Year()))
					monthSelect.SetSelected(parsed.Month().String())
				}
			}
			if editChart.RelocatedLatitude != "" && editChart.RelocatedLongitude != "" {
				relocateCheck.SetChecked(true)
				relocateLocation.SetText(editChart.RelocatedLocationName)
				relocateLat.SetText(editChart.RelocatedLatitude)
				relocateLon.SetText(editChart.RelocatedLongitude)
				relocateFields.Show()
			}
		} else {
			updateName()
		}

		cast := func() {
			base := lookup[baseSelect.Selected]
			natalTime, err := parseLocalDateTime(base.LocalDate, base.LocalTime)
			if err != nil {
				statusLabel.SetText("Error parsing natal birth time: " + err.Error())
				return
			}

			year, err := strconv.Atoi(yearEntry.Text)
			if err != nil || year < 1000 || year > 9999 {
				statusLabel.SetText("Invalid return year")
				return
			}

			monthIndex := 1
			for i, m := range months {
				if m == monthSelect.Selected {
					monthIndex = i + 1
					break
				}
			}

			var saved storage.SavedChart
			if isEdit {
				saved = *editChart
			} else {
				saved = base
				saved.ID = ""
			}
			saved.Name = nameEntry.Text
			saved.ChartType = string(astro.ChartTypeLunarReturn)
			saved.BaseChartID = base.ID
			saved.ComparisonChartID = ""
			saved.ReferenceDate = fmt.Sprintf("%04d-%02d-%02d", year, monthIndex, natalTime.Day())
			saved.ReferenceTime = formatClock(natalTime)
			saved.ReferenceUTC = time.Date(year, time.Month(monthIndex), natalTime.Day(), natalTime.Hour(), natalTime.Minute(), natalTime.Second(), 0, time.UTC).Format(time.RFC3339)
			saved.DirectionKey = ""

			saved.RelocatedLatitude = ""
			saved.RelocatedLongitude = ""
			saved.RelocatedLocationName = ""
			if relocateCheck.Checked {
				saved.RelocatedLatitude = relocateLat.Text
				saved.RelocatedLongitude = relocateLon.Text
				saved.RelocatedLocationName = relocateLocation.Text
			}

			if saveErr := store.Save(&saved); saveErr != nil {
				statusLabel.SetText(saveErr.Error())
				return
			}
			dataWindow.Close()
			selectedSavedID = saved.ID
			refreshSavedCharts()
			if chartList != nil {
				if selectedIndex := savedChartIndexByID(savedCharts, selectedSavedID); selectedIndex >= 0 {
					chartList.Select(selectedIndex)
				}
			}
			loadSavedChart(saved)
			status.SetText(fmt.Sprintf("Saved lunar return chart %s", saved.Name))
		}

		form := widget.NewForm(
			widget.NewFormItem("Base Natal Chart", baseSelect),
			widget.NewFormItem("Chart Name", nameEntry),
			widget.NewFormItem("Return Year", yearEntry),
			widget.NewFormItem("Return Month", monthSelect),
		)

		buttons := container.NewGridWithColumns(2,
			widget.NewButton("Cancel", dataWindow.Close),
			widget.NewButton(actionText, cast),
		)

		content := container.NewVBox(form, relocateCheck, relocateFields, statusLabel, buttons)
		dataWindow.SetContent(widget.NewCard(titleText, "Configure lunar return parameters.", content))
		dataWindow.Show()
	}

	showProgressionDialog := func(progType astro.ChartType, editChart *storage.SavedChart) {
		isEdit := editChart != nil
		actionText := "Cast"
		if isEdit {
			actionText = "Save"
		}

		titleText := ""
		switch progType {
		case astro.ChartTypeSecondaryProgression:
			titleText = "Secondary Progression"
		case astro.ChartTypeTertiaryProgression:
			titleText = "Tertiary Progression"
		case astro.ChartTypeSolarArc:
			titleText = "Solar Arc Chart"
		case astro.ChartTypePrimaryDirection:
			titleText = "Primary Directions Chart"
		}

		if isEdit {
			titleText = "Edit " + titleText
		} else {
			titleText = "Cast " + titleText
		}

		dataWindow := application.NewWindow(titleText)
		dataWindow.Resize(fyne.NewSize(450, 380))

		options, lookup := getNatalCharts()
		if len(options) == 0 {
			dialog.ShowInformation("No Natal Charts", "Please create a natal chart first.", window)
			return
		}

		baseSelect := createNatalSelector(options, lookup)

		nameEntry := widget.NewEntry()
		updateName := func() {
			if !isEdit {
				prefix := ""
				switch progType {
				case astro.ChartTypeSecondaryProgression:
					prefix = "Sec. Prog"
				case astro.ChartTypeTertiaryProgression:
					prefix = "Tert. Prog"
				case astro.ChartTypeSolarArc:
					prefix = "Solar Arc"
				case astro.ChartTypePrimaryDirection:
					prefix = "Prim. Dir"
				}
				nameEntry.SetText(fmt.Sprintf("%s - %s", prefix, lookup[baseSelect.Selected].Name))
			}
		}
		baseSelect.OnChanged = func(value string) { updateName() }

		targetMode := widget.NewSelect([]string{"Date", "Age"}, nil)
		targetMode.SetSelected("Date")

		dateEntry := widget.NewEntry()
		dateEntry.SetText(time.Now().Format("2006-01-02"))
		timeEntry := widget.NewEntry()
		timeEntry.SetText("12:00")

		ageEntry := widget.NewEntry()
		ageEntry.SetText("30")

		primaryKeySelect := widget.NewSelect([]string{"Naibod (59'08\"/yr)", "Ptolemy (1°/yr)"}, nil)
		primaryKeySelect.SetSelected("Naibod (59'08\"/yr)")

		statusLabel := widget.NewLabel("")

		dateFormItem1 := widget.NewFormItem("Target Date (YYYY-MM-DD)", dateEntry)
		dateFormItem2 := widget.NewFormItem("Target Time (HH:MM)", timeEntry)
		ageFormItem := widget.NewFormItem("Age (years)", ageEntry)
		keyFormItem := widget.NewFormItem("Direction Key", primaryKeySelect)

		form := widget.NewForm(
			widget.NewFormItem("Base Natal Chart", baseSelect),
			widget.NewFormItem("Chart Name", nameEntry),
			widget.NewFormItem("Target Mode", targetMode),
			dateFormItem1,
			dateFormItem2,
			ageFormItem,
		)

		if progType == astro.ChartTypePrimaryDirection {
			form.AppendItem(keyFormItem)
		}

		refreshFields := func() {
			if targetMode.Selected == "Age" {
				dateFormItem1.Widget.Hide()
				dateFormItem2.Widget.Hide()
				ageFormItem.Widget.Show()
			} else {
				dateFormItem1.Widget.Show()
				dateFormItem2.Widget.Show()
				ageFormItem.Widget.Hide()
			}
			form.Refresh()
		}
		targetMode.OnChanged = func(value string) {
			refreshFields()
		}

		if isEdit {
			nameEntry.SetText(editChart.Name)
			if baseChart, ok := savedChartByID(savedCharts, editChart.BaseChartID); ok {
				baseSelect.SetSelected(savedChartLabel(baseChart))
			}
			if editChart.ReferenceDate != "" {
				dateEntry.SetText(editChart.ReferenceDate)
				timeEntry.SetText(editChart.ReferenceTime)
			}
			if editChart.DirectionKey != "" {
				if editChart.DirectionKey == "ptolemy" {
					primaryKeySelect.SetSelected("Ptolemy (1°/yr)")
				} else {
					primaryKeySelect.SetSelected("Naibod (59'08\"/yr)")
				}
			}
		} else {
			updateName()
		}

		refreshFields()

		cast := func() {
			base := lookup[baseSelect.Selected]
			natalTime, err := parseLocalDateTime(base.LocalDate, base.LocalTime)
			if err != nil {
				statusLabel.SetText("Error parsing natal birth time: " + err.Error())
				return
			}

			var saved storage.SavedChart
			if isEdit {
				saved = *editChart
			} else {
				saved = base
				saved.ID = ""
			}
			saved.Name = nameEntry.Text
			saved.ChartType = string(progType)
			saved.BaseChartID = base.ID
			saved.ComparisonChartID = ""
			saved.RelocatedLatitude = ""
			saved.RelocatedLongitude = ""
			saved.RelocatedLocationName = ""
			saved.DirectionKey = ""

			if progType == astro.ChartTypePrimaryDirection {
				if primaryKeySelect.Selected == "Ptolemy (1°/yr)" {
					saved.DirectionKey = "ptolemy"
				} else {
					saved.DirectionKey = "naibod"
				}
			}

			if targetMode.Selected == "Age" {
				age, err := strconv.ParseFloat(ageEntry.Text, 64)
				if err != nil || age < 0 {
					statusLabel.SetText("Age must be a positive number")
					return
				}
				targetTime := natalTime.Add(time.Duration(age * 365.242199 * 24 * float64(time.Hour)))
				saved.ReferenceDate = targetTime.Format("2006-01-02")
				saved.ReferenceTime = formatClock(targetTime)
				saved.ReferenceUTC = targetTime.Format(time.RFC3339)
			} else {
				referenceUTC, err := parseReferenceDateTime(dateEntry.Text, timeEntry.Text)
				if err != nil {
					statusLabel.SetText(err.Error())
					return
				}
				saved.ReferenceDate = dateEntry.Text
				saved.ReferenceTime = timeEntry.Text
				saved.ReferenceUTC = referenceUTC.Format(time.RFC3339)
			}

			if saveErr := store.Save(&saved); saveErr != nil {
				statusLabel.SetText(saveErr.Error())
				return
			}
			dataWindow.Close()
			selectedSavedID = saved.ID
			refreshSavedCharts()
			if chartList != nil {
				if selectedIndex := savedChartIndexByID(savedCharts, selectedSavedID); selectedIndex >= 0 {
					chartList.Select(selectedIndex)
				}
			}
			loadSavedChart(saved)
			status.SetText(fmt.Sprintf("Saved progression chart %s", saved.Name))
		}

		buttons := container.NewGridWithColumns(2,
			widget.NewButton("Cancel", dataWindow.Close),
			widget.NewButton(actionText, cast),
		)

		content := container.NewBorder(nil, container.NewVBox(statusLabel, buttons), nil, nil, form)
		dataWindow.SetContent(widget.NewCard(titleText, "Configure progression or direction parameters.", content))
		dataWindow.Show()
	}

	showRelationshipDialog := func(relType astro.ChartType, editChart *storage.SavedChart) {
		isEdit := editChart != nil
		actionText := "Cast"
		if isEdit {
			actionText = "Save"
		}

		titleText := ""
		switch relType {
		case astro.ChartTypeComposite:
			titleText = "Composite Chart"
		case astro.ChartTypeDavison:
			titleText = "Davison Midpoint Chart"
		}

		if isEdit {
			titleText = "Edit " + titleText
		} else {
			titleText = "Cast " + titleText
		}

		dataWindow := application.NewWindow(titleText)
		dataWindow.Resize(fyne.NewSize(450, 260))

		options, lookup := getNatalCharts()
		if len(options) < 2 {
			dialog.ShowInformation("Insufficient Charts", "Please create at least two natal charts in your library to cast a relationship chart.", window)
			return
		}

		baseSelect := createNatalSelector(options, lookup)

		comparisonSelect := widget.NewSelect(nil, nil)

		refreshComparisonOptions := func() {
			compOpts := []string{}
			for _, opt := range options {
				if opt != baseSelect.Selected {
					compOpts = append(compOpts, opt)
				}
			}
			comparisonSelect.Options = compOpts
			comparisonSelect.Refresh()
			if len(compOpts) > 0 {
				if comparisonSelect.Selected == "" || comparisonSelect.Selected == baseSelect.Selected {
					comparisonSelect.SetSelected(compOpts[0])
				}
			}
		}

		nameEntry := widget.NewEntry()
		updateName := func() {
			if !isEdit {
				prefix := "Composite"
				if relType == astro.ChartTypeDavison {
					prefix = "Davison"
				}
				baseName := ""
				compName := ""
				if b, ok := lookup[baseSelect.Selected]; ok {
					baseName = b.Name
				}
				if c, ok := lookup[comparisonSelect.Selected]; ok {
					compName = c.Name
				}
				nameEntry.SetText(fmt.Sprintf("%s - %s / %s", prefix, baseName, compName))
			}
		}

		baseSelect.OnChanged = func(value string) {
			refreshComparisonOptions()
			updateName()
		}
		comparisonSelect.OnChanged = func(value string) {
			updateName()
		}

		refreshComparisonOptions()

		if isEdit {
			nameEntry.SetText(editChart.Name)
			if baseChart, ok := savedChartByID(savedCharts, editChart.BaseChartID); ok {
				baseSelect.SetSelected(savedChartLabel(baseChart))
			}
			refreshComparisonOptions()
			if comparisonChart, ok := savedChartByID(savedCharts, editChart.ComparisonChartID); ok {
				comparisonSelect.SetSelected(savedChartLabel(comparisonChart))
			}
		} else {
			updateName()
		}

		statusLabel := widget.NewLabel("")

		cast := func() {
			base := lookup[baseSelect.Selected]
			comp, ok := lookup[comparisonSelect.Selected]
			if !ok {
				statusLabel.SetText("Please select a comparison chart")
				return
			}
			if base.ID == comp.ID {
				statusLabel.SetText("Base and Comparison charts must be different")
				return
			}

			var saved storage.SavedChart
			if isEdit {
				saved = *editChart
			} else {
				saved = base
				saved.ID = ""
			}
			saved.Name = nameEntry.Text
			saved.ChartType = string(relType)
			saved.BaseChartID = base.ID
			saved.ComparisonChartID = comp.ID
			saved.ReferenceDate = ""
			saved.ReferenceTime = ""
			saved.ReferenceUTC = ""
			saved.RelocatedLatitude = ""
			saved.RelocatedLongitude = ""
			saved.RelocatedLocationName = ""
			saved.DirectionKey = ""

			if saveErr := store.Save(&saved); saveErr != nil {
				statusLabel.SetText(saveErr.Error())
				return
			}
			dataWindow.Close()
			selectedSavedID = saved.ID
			refreshSavedCharts()
			if chartList != nil {
				if selectedIndex := savedChartIndexByID(savedCharts, selectedSavedID); selectedIndex >= 0 {
					chartList.Select(selectedIndex)
				}
			}
			loadSavedChart(saved)
			status.SetText(fmt.Sprintf("Saved relationship chart %s", saved.Name))
		}

		form := widget.NewForm(
			widget.NewFormItem("First Natal Chart", baseSelect),
			widget.NewFormItem("Second Natal Chart", comparisonSelect),
			widget.NewFormItem("Chart Name", nameEntry),
		)

		buttons := container.NewGridWithColumns(2,
			widget.NewButton("Cancel", dataWindow.Close),
			widget.NewButton(actionText, cast),
		)

		content := container.NewBorder(nil, container.NewVBox(statusLabel, buttons), nil, nil, form)
		dataWindow.SetContent(widget.NewCard(titleText, "Select two charts to cast/edit their midpoint chart.", content))
		dataWindow.Show()
	}

	showEditDialog := func() {
		if !hasActiveSavedChart {
			status.SetText("Select a saved chart first")
			return
		}
		cType := astro.ChartType(activeSavedChart.ChartType)
		if cType == "" {
			cType = astro.ChartTypeNatal
		}

		switch cType {
		case astro.ChartTypeNatal:
			showBirthDataDialog()
		case astro.ChartTypeSolarReturn:
			showSolarReturnDialog(&activeSavedChart)
		case astro.ChartTypeLunarReturn:
			showLunarReturnDialog(&activeSavedChart)
		case astro.ChartTypeSecondaryProgression, astro.ChartTypeTertiaryProgression, astro.ChartTypeSolarArc, astro.ChartTypePrimaryDirection:
			showProgressionDialog(cType, &activeSavedChart)
		case astro.ChartTypeComposite, astro.ChartTypeDavison:
			showRelationshipDialog(cType, &activeSavedChart)
		}
	}

	showCompareDialog := func() {
		dataWindow := application.NewWindow("Compare / Bi-wheel")
		dataWindow.Resize(fyne.NewSize(400, 260))

		nextName := widget.NewEntry()
		nextName.SetText("New Bi-wheel")

		baseSelect := widget.NewSelect(nil, nil)
		comparisonSelect := widget.NewSelect(nil, nil)

		derivedLookup := map[string]storage.SavedChart{}
		options := []string{}
		for _, saved := range savedCharts {
			if saved.ChartType == string(astro.ChartTypeSynastry) {
				continue
			}
			label := savedChartLabel(saved)
			options = append(options, label)
			derivedLookup[label] = saved
		}
		baseSelect.Options = options
		comparisonSelect.Options = options

		if len(options) > 0 {
			if hasActiveSavedChart && activeSavedChart.ChartType != string(astro.ChartTypeSynastry) {
				baseSelect.SetSelected(savedChartLabel(activeSavedChart))
			} else {
				baseSelect.SetSelected(options[0])
			}
			if len(options) > 1 {
				comparisonSelect.SetSelected(options[1])
			} else {
				comparisonSelect.SetSelected(options[0])
			}
		}

		create := func() {
			base, ok := derivedLookup[baseSelect.Selected]
			if !ok {
				status.SetText("Select a base chart")
				return
			}
			comparison, ok := derivedLookup[comparisonSelect.Selected]
			if !ok {
				status.SetText("Select a comparison chart")
				return
			}
			if base.ID == comparison.ID {
				status.SetText("Choose two different charts for comparison")
				return
			}

			saved := base
			saved.ID = ""
			saved.Name = nextName.Text
			saved.ChartType = string(astro.ChartTypeSynastry)
			saved.BaseChartID = base.ID
			saved.ComparisonChartID = comparison.ID
			saved.ReferenceDate = ""
			saved.ReferenceTime = ""
			saved.ReferenceUTC = ""
			saved.RelocatedLatitude = ""
			saved.RelocatedLongitude = ""
			saved.RelocatedLocationName = ""
			saved.DirectionKey = ""

			if saveErr := store.Save(&saved); saveErr != nil {
				status.SetText(saveErr.Error())
				return
			}
			dataWindow.Close()
			selectedSavedID = saved.ID
			refreshSavedCharts()
			if chartList != nil {
				if selectedIndex := savedChartIndexByID(savedCharts, selectedSavedID); selectedIndex >= 0 {
					chartList.Select(selectedIndex)
				}
			}
			loadSavedChart(saved)
			status.SetText(fmt.Sprintf("Created Bi-wheel comparison %s", saved.Name))
		}

		buttons := container.NewGridWithColumns(2,
			widget.NewButton("Cancel", dataWindow.Close),
			widget.NewButton("Compare", create),
		)

		form := widget.NewForm(
			widget.NewFormItem("Name", nextName),
			widget.NewFormItem("Base Chart", baseSelect),
			widget.NewFormItem("Comparison Chart", comparisonSelect),
		)

		dataWindow.SetContent(widget.NewCard("Compare / Bi-wheel", "Select two charts to overlay in a bi-wheel.", container.NewBorder(
			nil,
			buttons,
			nil,
			nil,
			form,
		)))
		dataWindow.Show()
	}

	showBirthDataDialog = func() {
		dataWindow := application.NewWindow("Birth Data")
		dataWindow.Resize(fyne.NewSize(420, 360))

		editName := newTabbableEntry()
		editName.SetText(name.Text)
		editDate := newTabbableEntry()
		editDate.SetText(date.Text)
		editTime := newTabbableEntry()
		editTime.SetText(clock.Text)
		editOffset := newTabbableEntry()
		editOffset.SetText(timezone.Text)
		editLocation := newTabbableEntry()
		editLocation.SetText(locationName.Text)
		editLatitude := newTabbableEntry()
		editLatitude.SetText(latitude.Text)
		editLongitude := newTabbableEntry()
		editLongitude.SetText(longitude.Text)
		editHouseSystem := widget.NewSelect(astro.HouseSystemOptions(), nil)
		editHouseSystem.SetSelected(houseSystem.Selected)
		lookupLocation := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
			result, err := geocoder.Lookup(editLocation.Text)
			if err != nil {
				dialog.ShowError(err, dataWindow)
				return
			}
			editLocation.SetText(result.DisplayName)
			editLatitude.SetText(fmt.Sprintf("%.6f", result.Latitude))
			editLongitude.SetText(fmt.Sprintf("%.6f", result.Longitude))
			status.SetText(fmt.Sprintf("Resolved %s", result.DisplayName))
		})

		onCuspsChanged := func(string) {
			updateOffset(&editDate.Entry, &editTime.Entry, &editLatitude.Entry, &editLongitude.Entry, &editOffset.Entry, status)
		}
		editDate.OnChanged = onCuspsChanged
		editTime.OnChanged = onCuspsChanged
		editLatitude.OnChanged = onCuspsChanged
		editLongitude.OnChanged = onCuspsChanged

		updateOffset(&editDate.Entry, &editTime.Entry, &editLatitude.Entry, &editLongitude.Entry, &editOffset.Entry, status)

		apply := func() {
			name.SetText(editName.Text)
			date.SetText(editDate.Text)
			clock.SetText(editTime.Text)
			timezone.SetText(editOffset.Text)
			locationName.SetText(editLocation.Text)
			latitude.SetText(editLatitude.Text)
			longitude.SetText(editLongitude.Text)
			houseSystem.SetSelected(editHouseSystem.Selected)
			dataWindow.Close()
			if calculateActiveChart() {
				saveChart(selectedSavedID != "")
			}
		}

		onSubmit := func(string) {
			apply()
		}
		editName.OnSubmitted = onSubmit
		editDate.OnSubmitted = onSubmit
		editTime.OnSubmitted = onSubmit
		editOffset.OnSubmitted = onSubmit
		editLocation.OnSubmitted = onSubmit
		editLatitude.OnSubmitted = onSubmit
		editLongitude.OnSubmitted = onSubmit

		createLabeledField := func(labelText string, input fyne.CanvasObject) fyne.CanvasObject {
			lbl := widget.NewLabelWithStyle(labelText, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			return container.NewVBox(lbl, input)
		}

		nameField := createLabeledField("Name", editName)
		dateField := createLabeledField("Date", editDate)
		timeField := createLabeledField("Time", editTime)
		offsetField := createLabeledField("UTC Offset", editOffset)
		locationField := createLabeledField("Location", container.NewBorder(nil, nil, nil, lookupLocation, editLocation))
		latitudeField := createLabeledField("Latitude", editLatitude)
		longitudeField := createLabeledField("Longitude", editLongitude)
		housesField := createLabeledField("Houses", editHouseSystem)

		formLayout := container.NewVBox(
			nameField,
			container.NewGridWithColumns(2, dateField, timeField),
			locationField,
			container.NewGridWithColumns(3, latitudeField, longitudeField, offsetField),
			housesField,
		)

		buttons := container.NewGridWithColumns(2,
			widget.NewButton("Cancel", dataWindow.Close),
			widget.NewButton("Apply", apply),
		)

		dataWindow.SetContent(widget.NewCard("Birth Data", "Edit the active chart input.", container.NewBorder(
			nil,
			buttons,
			nil,
			nil,
			formLayout,
		)))
		dataWindow.Show()
	}

	setDarkTheme := func() {
		// nolint:staticcheck // Intentionally setting a theme variant on user action
		application.Settings().SetTheme(theme.DarkTheme())
		if activeSynastryChart != nil {
			refreshSynastryChart(*activeSynastryChart)
			return
		}
		refreshChart(currentChart)
	}
	setLightTheme := func() {
		// nolint:staticcheck // Intentionally setting a theme variant on user action
		application.Settings().SetTheme(theme.LightTheme())
		if activeSynastryChart != nil {
			refreshSynastryChart(*activeSynastryChart)
			return
		}
		refreshChart(currentChart)
	}

	showSwephInfo := func() {
		result, err := sweph.Smoke()
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		dialog.ShowInformation("Swiss Ephemeris", fmt.Sprintf(
			"Version: %s\nJulian day: %.5f\nSun longitude: %.6f",
			result.Version,
			result.JulianDay,
			result.SunLongitude,
		), window)
	}

	showGlobalSettings := func() {
		showSettingsDialog(window, store, func() {
			enabledChartObjects()
			if hasActiveSavedChart {
				resolvedChart, err := resolver.Resolve(activeSavedChart, savedCharts)
				if err != nil {
					status.SetText(err.Error())
					return
				}
				if resolvedChart.Synastry != nil {
					nextSynastry := *resolvedChart.Synastry
					if synastrySwapped {
						nextSynastry = swapSynastry(nextSynastry)
					}
					refreshSynastryChart(nextSynastry)
					return
				}
				if resolvedChart.Single != nil {
					refreshChart(*resolvedChart.Single)
					return
				}
			}
			calculateActiveChart()
		})
	}

	// Return submenu
	returnSubmenu := fyne.NewMenu("Return Charts",
		fyne.NewMenuItem("Solar Return", func() { showSolarReturnDialog(nil) }),
		fyne.NewMenuItem("Lunar Return", func() { showLunarReturnDialog(nil) }),
	)
	returnMenuItem := fyne.NewMenuItem("New Return Chart", nil)
	returnMenuItem.ChildMenu = returnSubmenu

	// Progression submenu
	progressionSubmenu := fyne.NewMenu("Progression / Direction Charts",
		fyne.NewMenuItem("Secondary Progression", func() { showProgressionDialog(astro.ChartTypeSecondaryProgression, nil) }),
		fyne.NewMenuItem("Tertiary Progression", func() { showProgressionDialog(astro.ChartTypeTertiaryProgression, nil) }),
		fyne.NewMenuItem("Solar Arc", func() { showProgressionDialog(astro.ChartTypeSolarArc, nil) }),
		fyne.NewMenuItem("Primary Directions", func() { showProgressionDialog(astro.ChartTypePrimaryDirection, nil) }),
	)
	progressionMenuItem := fyne.NewMenuItem("New Progression / Direction Chart", nil)
	progressionMenuItem.ChildMenu = progressionSubmenu

	// Relationship submenu
	relationshipSubmenu := fyne.NewMenu("Relationship Charts",
		fyne.NewMenuItem("Composite", func() { showRelationshipDialog(astro.ChartTypeComposite, nil) }),
		fyne.NewMenuItem("Davison", func() { showRelationshipDialog(astro.ChartTypeDavison, nil) }),
	)
	relationshipMenuItem := fyne.NewMenuItem("New Relationship Chart", nil)
	relationshipMenuItem.ChildMenu = relationshipSubmenu

	window.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("New Natal Chart", showNewNatalChartDialog),
			returnMenuItem,
			progressionMenuItem,
			relationshipMenuItem,
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Compare / Bi-wheel...", showCompareDialog),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", application.Quit),
		),
		fyne.NewMenu("View",
			fyne.NewMenuItem("Settings", showGlobalSettings),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Dark Mode", setDarkTheme),
			fyne.NewMenuItem("Light Mode", setLightTheme),
		),
		fyne.NewMenu("Tools",
			fyne.NewMenuItem("Swiss Ephemeris Smoke Test", showSwephInfo),
		),
	))

	chartList = widget.NewList(
		func() int {
			if len(savedCharts) == 0 {
				return 1
			}
			return len(savedCharts)
		},
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabelWithStyle("Chart name", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel("Date, time, location"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			row := obj.(*fyne.Container)
			title := row.Objects[0].(*widget.Label)
			meta := row.Objects[1].(*widget.Label)
			if len(savedCharts) == 0 {
				title.SetText("No saved charts")
				meta.SetText("Save the active chart to build a library.")
				return
			}
			saved := savedCharts[id]
			title.SetText(savedChartTitle(saved))
			meta.SetText(savedChartMeta(saved, savedCharts))
		},
	)
	chartList.OnSelected = func(id widget.ListItemID) {
		if len(savedCharts) == 0 {
			chartList.Unselect(id)
			selectedSavedID = ""
			status.SetText("No saved charts")
			return
		}
		if id < 0 || id >= len(savedCharts) {
			return
		}
		loadSavedChart(savedCharts[id])
	}
	if selectedSavedID != "" {
		if selectedIndex := savedChartIndexByID(savedCharts, selectedSavedID); selectedIndex >= 0 {
			chartList.Select(selectedIndex)
		}
	}

	deleteLibraryButton := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		if selectedSavedID == "" {
			status.SetText("Select a saved chart first")
			return
		}
		saved, ok := savedChartByID(savedCharts, selectedSavedID)
		if !ok {
			status.SetText("Select a saved chart first")
			return
		}
		dialog.ShowConfirm("Delete Saved Chart", fmt.Sprintf("Delete %s?", saved.Name), func(ok bool) {
			if ok {
				deleteSavedChart()
			}
		}, window)
	})
	editButton = widget.NewButtonWithIcon("Edit", theme.DocumentCreateIcon(), showEditDialog)
	savedPanel := widget.NewCard("Chart Library", "Select a saved chart to load it.", container.NewBorder(
		nil,
		container.NewGridWithColumns(3,
			editButton,
			widget.NewButtonWithIcon("Compare", theme.VisibilityIcon(), showCompareDialog),
			deleteLibraryButton,
		),
		nil,
		nil,
		chartList,
	))
	formPanel := container.NewBorder(
		nil,
		nil,
		nil,
		nil,
		savedPanel,
	)

	themeSelect := widget.NewSelect([]string{"Dark", "Light"}, func(value string) {
		if value == "Light" {
			setLightTheme()
			return
		}
		setDarkTheme()
	})
	themeSelect.SetSelected("Dark")

	backButton = widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() { navigateTime(-1) })
	forwardButton = widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() { navigateTime(1) })
	stepAmount = widget.NewEntry()
	stepAmount.SetText("1")
	stepAmount.SetPlaceHolder("1")
	stepAmount.Resize(fyne.NewSize(56, stepAmount.MinSize().Height))
	stepUnit = widget.NewSelect(timeStepUnits(), nil)
	stepUnit.SetSelected(timeStepUnitMinute)
	swapButton = widget.NewButtonWithIcon("Swap", theme.ViewRefreshIcon(), swapActiveSynastry)
	activeTimeLabel = widget.NewLabel("")
	workspaceLabel = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	timeToolbar := container.NewHBox(
		backButton,
		stepAmount,
		stepUnit,
		forwardButton,
		widget.NewSeparator(),
		activeTimeLabel,
		swapButton,
	)

	header := container.NewBorder(
		nil,
		nil,
		workspaceLabel,
		themeSelect,
		timeToolbar,
	)
	updateWorkspaceChrome()

	chartPanel := container.NewBorder(header, status, nil, nil, chartBody)
	mainSplit := container.NewHSplit(formPanel, chartPanel)
	mainSplit.Offset = 0.26

	window.SetContent(mainSplit)
	window.ShowAndRun()
}

func currentChartInputSummary(name, date, clock string) string {
	return fmt.Sprintf("%s\n%s %s", name, date, clock)
}

func derivedChartTypeOptions() []string {
	options := []string{}
	for _, chartType := range astro.SupportedChartTypes() {
		if chartType.SupportsDirectBirthData() {
			continue
		}
		options = append(options, chartType.String())
	}
	return options
}

func chartTypeFromLabel(value string) astro.ChartType {
	for _, chartType := range astro.SupportedChartTypes() {
		if chartType.String() == value {
			return chartType
		}
	}
	return astro.ChartTypeNatal
}

func chartTypeLabel(value string) string {
	return chartTypeFromLabel(value).String()
}

func defaultChartName(chartType astro.ChartType) string {
	switch chartType {
	case astro.ChartTypeTransit:
		return "New Transit Chart"
	case astro.ChartTypeSynastry:
		return "New Synastry Chart"
	case astro.ChartTypeSecondaryProgression:
		return "New Secondary Progression"
	case astro.ChartTypeSolarArc:
		return "New Solar Arc Chart"
	case astro.ChartTypeSolarReturn:
		return "New Solar Return"
	case astro.ChartTypeLunarReturn:
		return "New Lunar Return"
	default:
		return "New Natal Chart"
	}
}

func canCalculateSavedChart(chart storage.SavedChart) bool {
	return chart.ChartType == "" || chart.ChartType == string(astro.ChartTypeNatal)
}

func savedChartTitle(chart storage.SavedChart) string {
	if canCalculateSavedChart(chart) {
		return chart.Name
	}
	return fmt.Sprintf("%s: %s", chartTypeLabel(chart.ChartType), chart.Name)
}

func savedChartLabel(chart storage.SavedChart) string {
	return fmt.Sprintf("%s  %s %s", chart.Name, chart.LocalDate, chart.LocalTime)
}

func savedChartMeta(chart storage.SavedChart, charts []storage.SavedChart) string {
	if canCalculateSavedChart(chart) {
		location := shortenLocationName(chart.LocationName)
		if location == "" {
			location = fmt.Sprintf("%s, %s", chart.LatitudeDegrees, chart.LongitudeDegrees)
		}
		return fmt.Sprintf("%s %s UTC%s  %s  %s", chart.LocalDate, chart.LocalTime, chart.UTCOffset, location, astro.HouseSystemFromCode(chart.HouseSystem).Label())
	}
	baseName := chart.Name
	if base, ok := savedChartByID(charts, chart.BaseChartID); ok {
		baseName = base.Name
	}
	switch chart.ChartType {
	case string(astro.ChartTypeSynastry):
		comparisonName := ""
		if comparison, ok := savedChartByID(charts, chart.ComparisonChartID); ok {
			comparisonName = comparison.Name
		}
		return fmt.Sprintf("%s x %s", baseName, comparisonName)
	case string(astro.ChartTypeTransit), string(astro.ChartTypeSecondaryProgression), string(astro.ChartTypeSolarArc), string(astro.ChartTypeSolarReturn), string(astro.ChartTypeLunarReturn):
		return fmt.Sprintf("%s @ %s %s UTC", baseName, chart.ReferenceDate, chart.ReferenceTime)
	default:
		return fmt.Sprintf("%s definition", chartTypeLabel(chart.ChartType))
	}
}

func savedChartByID(charts []storage.SavedChart, id string) (storage.SavedChart, bool) {
	for _, chart := range charts {
		if chart.ID == id {
			return chart, true
		}
	}
	return storage.SavedChart{}, false
}

func savedChartIndexByID(charts []storage.SavedChart, id string) int {
	for i, chart := range charts {
		if chart.ID == id {
			return i
		}
	}
	return -1
}

func parseReferenceDateTime(date, clock string) (time.Time, error) {
	reference, err := parseLocalDateTime(date, clock)
	if err != nil {
		return time.Time{}, fmt.Errorf("reference date/time must use YYYY-MM-DD and HH:MM or HH:MM:SS")
	}
	return time.Date(reference.Year(), reference.Month(), reference.Day(), reference.Hour(), reference.Minute(), reference.Second(), 0, time.UTC), nil
}

func parseBirthData(name, date, clock, timezoneOffset, location, latitudeValue, longitudeValue, houseSystemLabel string) (astro.BirthData, error) {
	localTime, err := parseLocalDateTime(date, clock)
	if err != nil {
		return astro.BirthData{}, fmt.Errorf("date/time must use YYYY-MM-DD and HH:MM or HH:MM:SS")
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
	timeZone := time.FixedZone("chart", offsetSeconds)
	local := time.Date(
		localTime.Year(),
		localTime.Month(),
		localTime.Day(),
		localTime.Hour(),
		localTime.Minute(),
		localTime.Second(),
		0,
		timeZone,
	)

	tzName := timezone.LookupTimezone(latitude, longitude)

	return astro.BirthData{
		Name:             name,
		DateTimeUTC:      local.UTC(),
		LocationName:     location,
		LatitudeDegrees:  latitude,
		LongitudeDegrees: longitude,
		HouseSystem:      astro.HouseSystemFromLabel(houseSystemLabel),
		UTCOffset:        timezoneOffset,
		TimezoneName:     tzName,
		ChartType:        astro.ChartTypeNatal,
	}, nil
}

func updateOffset(dateEntry, timeEntry, latEntry, lonEntry, offsetEntry *widget.Entry, statusLabel *widget.Label) {
	lat, err1 := strconv.ParseFloat(latEntry.Text, 64)
	lon, err2 := strconv.ParseFloat(lonEntry.Text, 64)
	if err1 != nil || err2 != nil {
		return // Ignore partial inputs while typing
	}
	tzName := timezone.LookupTimezone(lat, lon)
	if tzName == "" {
		return
	}
	offset, err := timezone.CalculateOffset(tzName, dateEntry.Text, timeEntry.Text)
	if err == nil {
		offsetEntry.SetText(fmt.Sprintf("%g", offset))
		if statusLabel != nil {
			statusLabel.SetText(fmt.Sprintf("Timezone: %s (Offset: %g)", tzName, offset))
		}
	}
}

type tabbableEntry struct {
	widget.Entry
}

func newTabbableEntry() *tabbableEntry {
	e := &tabbableEntry{}
	e.ExtendBaseWidget(e)
	return e
}

func (e *tabbableEntry) AcceptsTab() bool {
	return false
}
