package ui

import (
	"fmt"
	"image/color"
	"math"

	"astro-go/internal/assets"
	"astro-go/internal/astro"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func NewSynastryWheel(chart astro.SynastryChart) fyne.CanvasObject {
	wheel := &synastryWheel{chart: chart}
	wheel.ExtendBaseWidget(wheel)
	return wheel
}

type synastryWheel struct {
	widget.BaseWidget

	chart   astro.SynastryChart
	content *fyne.Container
}

func (w *synastryWheel) CreateRenderer() fyne.WidgetRenderer {
	w.content = container.NewWithoutLayout()
	renderer := &synastryWheelRenderer{wheel: w}
	w.layout(w.Size())
	return renderer
}

func (w *synastryWheel) layout(size fyne.Size) {
	if w.content == nil {
		return
	}
	w.content.Resize(size)
	w.content.Objects = synastryWheelObjects(w.chart, size)
	w.content.Refresh()
}

type synastryWheelRenderer struct {
	wheel *synastryWheel
}

func (r *synastryWheelRenderer) Layout(size fyne.Size) {
	r.wheel.layout(size)
}

func (r *synastryWheelRenderer) MinSize() fyne.Size {
	return fyne.NewSize(220, 220)
}

func (r *synastryWheelRenderer) Refresh() {
	r.wheel.layout(r.wheel.Size())
}

func (r *synastryWheelRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.wheel.content}
}

func (r *synastryWheelRenderer) Destroy() {}

func synastryWheelObjects(chart astro.SynastryChart, size fyne.Size) []fyne.CanvasObject {
	palette := currentChartPalette()
	wheelAreaWidth := float64(size.Width)
	drawSize := math.Min(wheelAreaWidth, float64(size.Height)) - 24
	if drawSize < 220 {
		drawSize = 220
	}

	centerX := wheelAreaWidth * 0.5
	centerY := float64(size.Height) / 2
	outerZodiacOuter := drawSize * 0.49
	zodiacBandWidth := drawSize * 0.04
	houseBandWidth := drawSize * 0.028
	outerZodiacInner := outerZodiacOuter - zodiacBandWidth
	outerZodiacSignRadius := (outerZodiacOuter + outerZodiacInner) * 0.5
	outer := drawSize * 0.312
	middle := outer - zodiacBandWidth
	outerHouseInner := outer
	outerHouseOuter := outerHouseInner + houseBandWidth
	outerPlanetExactRadius := outerZodiacInner
	planetBandWidth := outerZodiacInner - outerHouseOuter
	planetLabelRadiusRatio := 0.58
	outerPlanetRadius := outerHouseOuter + planetBandWidth*planetLabelRadiusRatio
	innerHouseOuter := middle - planetBandWidth - drawSize*0.025
	innerHouseInner := innerHouseOuter - houseBandWidth
	aspectRadius := innerHouseInner
	innerPlanetExactRadius := middle
	innerPlanetBandWidth := middle - innerHouseOuter
	innerPlanetRadius := innerHouseOuter + innerPlanetBandWidth*planetLabelRadiusRatio
	signRadius := (outer + middle) * 0.5
	innerZodiacCuspDegreeRadius := signRadius
	outerZodiacCuspDegreeRadius := outerZodiacSignRadius
	houseNumberRadius := innerHouseInner + (innerHouseOuter-innerHouseInner)*0.6
	outerHouseNumberRadius := outerHouseInner + (outerHouseOuter-outerHouseInner)*0.5
	ascendant := chart.InnerChart.Ascendant.Longitude

	signTextSize := float32(clamp(drawSize*0.032, 10, 18))
	cuspDegreeTextSize := float32(clamp(drawSize*0.017, 7, 10))
	planetTextSize := float32(clamp(drawSize*0.049, 15, 27))
	houseTextSize := float32(clamp(drawSize*0.024, 9, 13))
	aspectTextSize := float32(clamp(drawSize*0.018, 8, 11))
	coordTextSize := float32(clamp(drawSize*0.023, 10, 13))

	objects := []fyne.CanvasObject{
		background(size, palette.background),
		filledCircle(centerX, centerY, outerZodiacOuter, color.NRGBA{R: 246, G: 241, B: 229, A: 255}, color.Transparent, 0),
		filledCircle(centerX, centerY, outerZodiacInner, palette.background, color.Transparent, 0),
		filledCircle(centerX, centerY, outer, color.NRGBA{R: 246, G: 241, B: 229, A: 255}, color.Transparent, 0),
		filledCircle(centerX, centerY, middle, palette.background, color.Transparent, 0),
		circle(centerX, centerY, outerZodiacOuter, palette.wheel, 2),
		circle(centerX, centerY, outerZodiacInner, palette.wheel, 2),
		circle(centerX, centerY, outerHouseOuter, palette.wheel, 2),
		circle(centerX, centerY, outerHouseInner, palette.wheel, 2),
		circle(centerX, centerY, outer, palette.wheel, 2),
		circle(centerX, centerY, middle, palette.wheel, 1.6),
		circle(centerX, centerY, innerHouseOuter, palette.wheel, 2),
		circle(centerX, centerY, innerHouseInner, palette.wheel, 2),
		circle(centerX, centerY, aspectRadius, palette.subtle, 1),
	}

	for degree := 0; degree < 360; degree++ {
		longitude := float64(degree)
		markerLength := drawSize * 0.008
		width := float32(0.35)
		tickColor := palette.tick
		if degree%5 == 0 {
			markerLength = drawSize * 0.014
			width = 0.75
			tickColor = palette.subtle
		}
		if degree%30 == 0 {
			markerLength = drawSize * 0.02
			width = 1.25
			tickColor = palette.wheel
		}
		x1, y1 := chartPoint(centerX, centerY, outer-markerLength, longitude, ascendant)
		x2, y2 := chartPoint(centerX, centerY, outer, longitude, ascendant)
		objects = append(objects, line(x1, y1, x2, y2, tickColor, width))

		x3, y3 := chartPoint(centerX, centerY, middle, longitude, ascendant)
		x4, y4 := chartPoint(centerX, centerY, middle+markerLength*0.75, longitude, ascendant)
		objects = append(objects, line(x3, y3, x4, y4, tickColor, width))

		x5, y5 := chartPoint(centerX, centerY, outerZodiacInner, longitude, ascendant)
		x6, y6 := chartPoint(centerX, centerY, outerZodiacInner+markerLength*0.75, longitude, ascendant)
		objects = append(objects, line(x5, y5, x6, y6, tickColor, width))
	}

	for i := 0; i < 12; i++ {
		longitude := float64(i * 30)
		x1, y1 := chartPoint(centerX, centerY, innerHouseInner, longitude, ascendant)
		x2, y2 := chartPoint(centerX, centerY, outer, longitude, ascendant)
		objects = append(objects, line(x1, y1, x2, y2, palette.wheel, 1.3))

		ox1, oy1 := chartPoint(centerX, centerY, outerZodiacInner, longitude, ascendant)
		ox2, oy2 := chartPoint(centerX, centerY, outerZodiacOuter, longitude, ascendant)
		objects = append(objects, line(ox1, oy1, ox2, oy2, palette.wheel, 1.2))

		labelLong := longitude + 15
		x, y := chartPoint(centerX, centerY, signRadius, labelLong, ascendant)
		text := canvas.NewText(signGlyph(astro.Sign(i)), palette.sign)
		text.TextSize = signTextSize
		text.FontSource = astrologyFont()
		moveTextCentered(text, x, y)
		objects = append(objects, text)

		ox, oy := chartPoint(centerX, centerY, outerZodiacSignRadius, labelLong, ascendant)
		outerText := canvas.NewText(signGlyph(astro.Sign(i)), palette.sign)
		outerText.TextSize = signTextSize
		outerText.FontSource = astrologyFont()
		moveTextCentered(outerText, ox, oy)
		objects = append(objects, outerText)
	}

	showHouseNumbers := len(chart.InnerChart.Houses) <= 18
	for index, house := range chart.InnerChart.Houses {
		longitude := house.CuspLongitude
		isAngularHouse := house.Number == 1 || house.Number == 4 || house.Number == 7 || house.Number == 10
		width := float32(1)
		if len(chart.InnerChart.Houses) == 12 && isAngularHouse {
			width = 3
		}
		x1, y1 := chartPoint(centerX, centerY, innerHouseInner, longitude, ascendant)
		houseLineOuterRadius := innerHouseOuter
		if isAngularHouse {
			houseLineOuterRadius = outerZodiacOuter
		}
		x2, y2 := chartPoint(centerX, centerY, houseLineOuterRadius, longitude, ascendant)
		objects = append(objects, line(x1, y1, x2, y2, palette.house, width))

		if showHouseNumbers {
			labelLong := houseLabelLongitude(chart.InnerChart.Houses, index)
			hx, hy := chartPoint(centerX, centerY, houseNumberRadius, labelLong, ascendant)
			houseText := canvas.NewText(fmt.Sprintf("%d", house.Number), palette.houseNumber)
			houseText.TextSize = houseTextSize
			houseText.Move(fyne.NewPos(float32(hx)-houseTextSize*0.32, float32(hy)-houseTextSize*0.58))
			objects = append(objects, houseText)
		}
	}

	showOuterHouseNumbers := len(chart.OuterChart.Houses) <= 18 && len(chart.OuterChart.Houses) > 0
	for index, house := range chart.OuterChart.Houses {
		longitude := house.CuspLongitude
		width := float32(1)
		if len(chart.OuterChart.Houses) == 12 && (house.Number == 1 || house.Number == 4 || house.Number == 7 || house.Number == 10) {
			width = 3
		}
		// Notice that the ascendant passed is the inner chart's ascendant because the whole wheel is rotated
		// based on the inner chart's ascendant!
		x1, y1 := chartPoint(centerX, centerY, outer, longitude, ascendant)
		x2, y2 := chartPoint(centerX, centerY, outerZodiacInner, longitude, ascendant)
		objects = append(objects, line(x1, y1, x2, y2, palette.house, width))

		if showOuterHouseNumbers {
			labelLong := houseLabelLongitude(chart.OuterChart.Houses, index)
			hx, hy := chartPoint(centerX, centerY, outerHouseNumberRadius, labelLong, ascendant)
			houseText := canvas.NewText(fmt.Sprintf("%d", house.Number), palette.houseNumber)
			houseText.TextSize = houseTextSize
			houseText.Move(fyne.NewPos(float32(hx)-houseTextSize*0.32, float32(hy)-houseTextSize*0.58))
			objects = append(objects, houseText)
		}
	}

	for _, house := range chart.InnerChart.Houses {
		objects = append(objects, cuspDegreeMinuteTexts(centerX, centerY, innerZodiacCuspDegreeRadius, house.CuspLongitude, ascendant, palette.houseNumber, cuspDegreeTextSize)...)
	}

	for _, house := range chart.OuterChart.Houses {
		objects = append(objects, cuspDegreeMinuteTexts(centerX, centerY, outerZodiacCuspDegreeRadius, house.CuspLongitude, ascendant, palette.houseNumber, cuspDegreeTextSize)...)
	}

	objects = append(objects, angularMarkers(centerX, centerY, outerZodiacOuter, innerHouseInner, ascendant, chart.InnerChart, palette, houseTextSize)...)
	objects = append(objects, angularMarkers(centerX, centerY, outerZodiacOuter, outer, ascendant, chart.OuterChart, palette, houseTextSize)...)

	for _, aspect := range chart.InterAspects {
		from, okFrom := planetLongitude(chart.InnerChart, aspect.Inner)
		to, okTo := planetLongitude(chart.OuterChart, aspect.Outer)
		if !okFrom || !okTo {
			continue
		}
		x1, y1 := chartPoint(centerX, centerY, aspectRadius, from, ascendant)
		x2, y2 := chartPoint(centerX, centerY, aspectRadius, to, ascendant)
		stroke := aspectColor(aspect.Type, palette)
		objects = append(objects, line(x1, y1, x2, y2, stroke, 1))
		midX, midY := (x1+x2)/2, (y1+y2)/2
		symbol := canvas.NewText(aspectGlyph(aspect.Type), stroke)
		symbol.TextSize = aspectTextSize
		symbol.FontSource = astrologyFont()
		moveTextCentered(symbol, midX, midY)
		objects = append(objects, symbol)
	}

	innerPlacements := planetPlacements(chart.InnerChart.Planets, centerX, centerY, innerPlanetExactRadius, innerPlanetRadius, ascendant, drawSize*0.025, float64(planetTextSize))
	outerPlacements := planetPlacements(chart.OuterChart.Planets, centerX, centerY, outerPlanetExactRadius, outerPlanetRadius, ascendant, drawSize*0.025, float64(planetTextSize))

	for _, placement := range innerPlacements {
		ex, ey := chartPoint(centerX, centerY, innerPlanetExactRadius, placement.position.Longitude, ascendant)
		lx, ly := planetLineEndpoint(centerX, centerY, placement.x, placement.y, float64(planetTextSize)*0.85)
		stroke := planetColor(placement.position.Planet, palette)
		objects = append(objects, line(ex, ey, lx, ly, stroke, 1.15))
	}
	for _, placement := range outerPlacements {
		ex, ey := chartPoint(centerX, centerY, outerPlanetExactRadius, placement.position.Longitude, ascendant)
		lx, ly := planetLineEndpoint(centerX, centerY, placement.x, placement.y, float64(planetTextSize)*0.85)
		stroke := withAlpha(planetColor(placement.position.Planet, palette), 245)
		objects = append(objects, line(ex, ey, lx, ly, stroke, 1.15))
	}
	for _, placement := range innerPlacements {
		ex, ey := chartPoint(centerX, centerY, innerPlanetExactRadius, placement.position.Longitude, ascendant)
		stroke := planetColor(placement.position.Planet, palette)
		objects = append(objects, filledCircle(ex, ey, drawSize*0.0045, stroke, stroke, 0.5))
		objects = append(objects, planetLabelObjects(placement.position, centerX, centerY, placement.x, placement.y, planetTextSize, coordTextSize, stroke)...)
	}
	for _, placement := range outerPlacements {
		ex, ey := chartPoint(centerX, centerY, outerPlanetExactRadius, placement.position.Longitude, ascendant)
		stroke := withAlpha(planetColor(placement.position.Planet, palette), 245)
		objects = append(objects, filledCircle(ex, ey, drawSize*0.0045, stroke, stroke, 0.5))
		objects = append(objects, planetLabelObjects(placement.position, centerX, centerY, placement.x, placement.y, planetTextSize, coordTextSize, stroke)...)
	}

	objects = append(objects, inCanvasSynastryInfo(chart, palette, size)...)

	return objects
}

func angularMarkers(centerX, centerY, markerOuterRadius, houseInner, ascendant float64, chart astro.Chart, palette chartPalette, houseTextSize float32) []fyne.CanvasObject {
	objects := []fyne.CanvasObject{}
	for _, marker := range []struct {
		label     string
		longitude float64
		color     color.Color
	}{
		{"ASC", chart.Ascendant.Longitude, palette.accent},
		{"DSC", astro.NormalizeDegrees(chart.Ascendant.Longitude + 180), palette.accent},
		{"MC", chart.MC.Longitude, palette.accent},
		{"IC", astro.NormalizeDegrees(chart.MC.Longitude + 180), palette.accent},
	} {
		x1, y1 := chartPoint(centerX, centerY, houseInner, marker.longitude, ascendant)
		x2, y2 := chartPoint(centerX, centerY, markerOuterRadius, marker.longitude, ascendant)
		objects = append(objects, line(x1, y1, x2, y2, marker.color, 3))
		labelOuterRadius := markerOuterRadius + float64(houseTextSize)*0.9
		lx, ly := chartPoint(centerX, centerY, labelOuterRadius, marker.longitude, ascendant)
		label := canvas.NewText(marker.label, marker.color)
		label.TextSize = houseTextSize
		label.FontSource = assets.CourierFont
		label.Move(angularMarkerLabelPosition(marker.label, lx, ly, houseTextSize))
		objects = append(objects, label)
	}
	return objects
}

func formatWheelDegreeMinutes(longitude float64) string {
	degrees, minutes := wheelDegreeMinuteParts(longitude)
	return fmt.Sprintf("%02d°%02d'", degrees, minutes)
}

func wheelDegreeMinuteParts(longitude float64) (int, int) {
	degreeInSign := astro.DegreeInSign(longitude)
	degrees := int(math.Floor(degreeInSign))
	minutes := int(math.Round((degreeInSign - float64(degrees)) * 60))
	if minutes == 60 {
		minutes = 0
		degrees++
	}
	if degrees == 30 {
		degrees = 29
		minutes = 59
	}
	return degrees, minutes
}

func cuspDegreeMinuteTexts(centerX, centerY, radius, longitude, ascendant float64, clr color.Color, size float32) []fyne.CanvasObject {
	degrees, minutes := wheelDegreeMinuteParts(longitude)
	degreeX, degreeY := chartPoint(centerX, centerY, radius, astro.NormalizeDegrees(longitude-1.1), ascendant)
	minuteX, minuteY := chartPoint(centerX, centerY, radius, astro.NormalizeDegrees(longitude+1.1), ascendant)
	return []fyne.CanvasObject{
		centeredText(fmt.Sprintf("%02d°", degrees), clr, size, degreeX, degreeY, assets.CourierFont),
		centeredText(fmt.Sprintf("%02d'", minutes), clr, size, minuteX, minuteY, assets.CourierFont),
	}
}

func centeredText(value string, clr color.Color, size float32, x, y float64, font fyne.Resource) fyne.CanvasObject {
	text := canvas.NewText(value, clr)
	text.TextSize = size
	text.FontSource = font
	text.Move(fyne.NewPos(float32(x)-float32(len(value))*size*0.3, float32(y)-size*0.5))
	return text
}

func inCanvasSynastryInfo(chart astro.SynastryChart, palette chartPalette, size fyne.Size) []fyne.CanvasObject {
	x := 12.0
	y := 18.0
	headerSize := float32(12)
	bodySize := float32(10)

	objects := []fyne.CanvasObject{
		textAt(chart.Name, palette.text, headerSize+2, x, y, true, assets.CourierFont),
	}
	y += 18
	objects = append(objects, textAt("Inner Chart", palette.text, bodySize, x, y, true, assets.CourierFont))
	y += 16
	objects = append(objects, textAt(chart.InnerChart.Name, palette.text, bodySize, x, y, true, assets.CourierFont))
	y += 16
	objects, y = appendChartInfoDetails(objects, chart.InnerChart, palette, x, y, bodySize)
	y += 12
	objects = append(objects, textAt("Outer Chart", palette.mutedText, bodySize, x, y, true, assets.CourierFont))
	y += 16
	objects = append(objects, textAt(chart.OuterChart.Name, palette.text, bodySize, x, y, true, assets.CourierFont))
	y += 16
	objects, _ = appendChartInfoDetails(objects, chart.OuterChart, palette, x, y, bodySize)

	return objects
}
