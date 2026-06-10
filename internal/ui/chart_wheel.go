package ui

import (
	"fmt"
	"image/color"
	"math"
	"sort"
	"strconv"
	"time"

	"astro-go/internal/assets"
	"astro-go/internal/astro"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewChartWheel(chart astro.Chart) fyne.CanvasObject {
	wheel := &chartWheel{chart: chart}
	wheel.ExtendBaseWidget(wheel)
	return wheel
}

type chartWheel struct {
	widget.BaseWidget

	chart   astro.Chart
	content *fyne.Container
}

func (w *chartWheel) CreateRenderer() fyne.WidgetRenderer {
	w.content = container.NewWithoutLayout()
	renderer := &chartWheelRenderer{wheel: w}
	w.layout(w.Size())
	return renderer
}

func (w *chartWheel) layout(size fyne.Size) {
	if w.content == nil {
		return
	}
	w.content.Resize(size)
	w.content.Objects = chartWheelObjects(w.chart, size)
	w.content.Refresh()
}

type chartWheelRenderer struct {
	wheel *chartWheel
}

func (r *chartWheelRenderer) Layout(size fyne.Size) {
	r.wheel.layout(size)
}

func (r *chartWheelRenderer) MinSize() fyne.Size {
	return fyne.NewSize(220, 220)
}

func (r *chartWheelRenderer) Refresh() {
	r.wheel.layout(r.wheel.Size())
}

func (r *chartWheelRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.wheel.content}
}

func (r *chartWheelRenderer) Destroy() {}

func chartWheelObjects(chart astro.Chart, size fyne.Size) []fyne.CanvasObject {
	palette := currentChartPalette()
	wheelAreaWidth := float64(size.Width)
	drawSize := math.Min(wheelAreaWidth, float64(size.Height)) - 24
	if drawSize < 180 {
		drawSize = 180
	}

	centerX := wheelAreaWidth * 0.5
	centerY := float64(size.Height) / 2
	outer := drawSize * 0.47
	zodiacInner := drawSize * 0.415
	planetExactRadius := zodiacInner
	planetRadius := zodiacInner - drawSize*0.05
	houseOuter := drawSize * 0.255
	houseInner := drawSize * 0.205
	houseNumberRadius := houseInner + (houseOuter-houseInner)*0.58
	aspectOuter := houseInner
	aspectRadius := aspectOuter
	signRadius := (outer + zodiacInner) * 0.5
	ascendant := chart.Ascendant.Longitude
	signTextSize := float32(clamp(drawSize*0.039, 12, 22))
	planetTextSize := float32(clamp(drawSize*0.049, 15, 27))
	coordTextSize := float32(clamp(drawSize*0.023, 10, 13))
	houseTextSize := float32(clamp(drawSize*0.024, 9, 13))
	aspectTextSize := float32(clamp(drawSize*0.018, 8, 11))

	objects := []fyne.CanvasObject{
		background(size, palette.background),
		filledCircle(centerX, centerY, outer, color.NRGBA{R: 246, G: 241, B: 229, A: 255}, color.Transparent, 0),
		filledCircle(centerX, centerY, zodiacInner, palette.background, color.Transparent, 0),
		circle(centerX, centerY, outer, palette.wheel, 2),
		circle(centerX, centerY, zodiacInner, palette.wheel, 2),
		circle(centerX, centerY, houseOuter, palette.wheel, 2),
		circle(centerX, centerY, houseInner, palette.wheel, 2),
		circle(centerX, centerY, aspectOuter, palette.subtle, 1),
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

		x3, y3 := chartPoint(centerX, centerY, zodiacInner, longitude, ascendant)
		x4, y4 := chartPoint(centerX, centerY, zodiacInner+markerLength*0.75, longitude, ascendant)
		objects = append(objects, line(x3, y3, x4, y4, tickColor, width))
	}

	for i := 0; i < 12; i++ {
		longitude := float64(i * 30)
		x1, y1 := chartPoint(centerX, centerY, zodiacInner, longitude, ascendant)
		x2, y2 := chartPoint(centerX, centerY, outer, longitude, ascendant)
		objects = append(objects, line(x1, y1, x2, y2, palette.wheel, 1.4))

		labelLong := longitude + 15
		x, y := chartPoint(centerX, centerY, signRadius, labelLong, ascendant)
		text := canvas.NewText(signGlyph(astro.Sign(i)), palette.sign)
		text.TextSize = signTextSize
		text.FontSource = astrologyFont()
		moveTextCentered(text, x, y)
		objects = append(objects, text)
	}

	showHouseNumbers := len(chart.Houses) <= 18
	for index, house := range chart.Houses {
		longitude := house.CuspLongitude
		width := float32(1)
		if len(chart.Houses) == 12 && (house.Number == 1 || house.Number == 4 || house.Number == 7 || house.Number == 10) {
			width = 3
		}
		x1, y1 := chartPoint(centerX, centerY, houseInner, longitude, ascendant)
		x2, y2 := chartPoint(centerX, centerY, zodiacInner, longitude, ascendant)
		objects = append(objects, line(x1, y1, x2, y2, palette.house, width))

		if showHouseNumbers {
			labelLong := houseLabelLongitude(chart.Houses, index)
			hx, hy := chartPoint(centerX, centerY, houseNumberRadius, labelLong, ascendant)
			houseText := canvas.NewText(fmt.Sprintf("%d", house.Number), palette.houseNumber)
			houseText.TextSize = houseTextSize
			houseText.Move(fyne.NewPos(float32(hx)-houseTextSize*0.32, float32(hy)-houseTextSize*0.58))
			objects = append(objects, houseText)
		}
	}

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
		lineOuterRadius := outer + drawSize*0.035
		labelOuterRadius := outer + drawSize*0.055
		x2, y2 := chartPoint(centerX, centerY, lineOuterRadius, marker.longitude, ascendant)
		objects = append(objects, line(x1, y1, x2, y2, marker.color, 3))
		lx, ly := chartPoint(centerX, centerY, labelOuterRadius, marker.longitude, ascendant)
		label := canvas.NewText(marker.label, marker.color)
		label.TextSize = houseTextSize
		label.FontSource = assets.CourierFont
		label.Move(angularMarkerLabelPosition(marker.label, lx, ly, houseTextSize))
		objects = append(objects, label)
	}

	for _, aspect := range chart.Aspects {
		from, okFrom := planetLongitude(chart, aspect.From)
		to, okTo := planetLongitude(chart, aspect.To)
		if !okFrom || !okTo {
			continue
		}
		x1, y1 := chartPoint(centerX, centerY, aspectRadius, from, ascendant)
		x2, y2 := chartPoint(centerX, centerY, aspectRadius, to, ascendant)
		aspectStroke := aspectColor(aspect.Type, palette)
		objects = append(objects, line(x1, y1, x2, y2, aspectStroke, 1))
		t1x1, t1y1 := chartPoint(centerX, centerY, aspectRadius-drawSize*0.01, from, ascendant)
		t1x2, t1y2 := chartPoint(centerX, centerY, aspectRadius+drawSize*0.01, from, ascendant)
		t2x1, t2y1 := chartPoint(centerX, centerY, aspectRadius-drawSize*0.01, to, ascendant)
		t2x2, t2y2 := chartPoint(centerX, centerY, aspectRadius+drawSize*0.01, to, ascendant)
		objects = append(objects, line(t1x1, t1y1, t1x2, t1y2, aspectStroke, 1.1))
		objects = append(objects, line(t2x1, t2y1, t2x2, t2y2, aspectStroke, 1.1))
		objects = append(objects, filledCircle(x1, y1, drawSize*0.0038, palette.background, aspectStroke, 0.8))
		objects = append(objects, filledCircle(x2, y2, drawSize*0.0038, palette.background, aspectStroke, 0.8))
		midX, midY := (x1+x2)/2, (y1+y2)/2
		symbol := canvas.NewText(aspectGlyph(aspect.Type), aspectStroke)
		symbol.TextSize = aspectTextSize
		symbol.FontSource = astrologyFont()
		moveTextCentered(symbol, midX, midY)
		objects = append(objects, symbol)
	}

	placements := planetPlacements(chart.Planets, centerX, centerY, planetExactRadius, planetRadius, ascendant, drawSize*0.025, float64(planetTextSize))
	for _, placement := range placements {
		ex, ey := chartPoint(centerX, centerY, planetExactRadius, placement.position.Longitude, ascendant)
		lx, ly := planetLineEndpoint(centerX, centerY, placement.x, placement.y, float64(planetTextSize)*0.85)
		stroke := planetColor(placement.position.Planet, palette)
		objects = append(objects, line(ex, ey, lx, ly, stroke, 1.15))
	}
	for _, placement := range placements {
		ex, ey := chartPoint(centerX, centerY, planetExactRadius, placement.position.Longitude, ascendant)
		stroke := planetColor(placement.position.Planet, palette)
		objects = append(objects, filledCircle(ex, ey, drawSize*0.0045, stroke, stroke, 0.5))
		objects = append(objects, planetLabelObjects(placement.position, centerX, centerY, placement.x, placement.y, planetTextSize, coordTextSize, stroke)...)
	}

	ascText := canvas.NewText(fmt.Sprintf("Asc %s", formatZodiacDMS(chart.Ascendant.Longitude)), palette.accent)
	ascText.TextSize = coordTextSize
	ascText.FontSource = assets.CourierFont
	ascText.Move(fyne.NewPos(float32(centerX-outer), float32(centerY+outer+float64(coordTextSize)*0.4)))
	objects = append(objects, ascText)

	mcText := canvas.NewText(fmt.Sprintf("MC %s", formatZodiacDMS(chart.MC.Longitude)), palette.accent)
	mcText.TextSize = coordTextSize
	mcText.FontSource = assets.CourierFont
	mcText.Move(fyne.NewPos(float32(centerX-outer), float32(centerY+outer+float64(coordTextSize)*1.9)))
	objects = append(objects, mcText)

	objects = append(objects, inCanvasChartInfo(chart, palette, size)...)

	return objects
}

type planetPlacement struct {
	position astro.PlanetPosition
	x        float64
	y        float64
}

type chartPalette struct {
	background      color.Color
	wheel           color.Color
	subtle          color.Color
	sign            color.Color
	planet          color.Color
	sun             color.Color
	moon            color.Color
	mercury         color.Color
	venus           color.Color
	mars            color.Color
	jupiter         color.Color
	saturn          color.Color
	house           color.Color
	houseNumber     color.Color
	tick            color.Color
	accent          color.Color
	secondaryAccent color.Color
	easyAspect      color.Color
	hardAspect      color.Color
	neutralAspect   color.Color
	text            color.Color
	mutedText       color.Color
}

func currentChartPalette() chartPalette {
	if fyne.CurrentApp() == nil {
		return darkChartPalette()
	}
	if fyne.CurrentApp().Settings().ThemeVariant() == theme.VariantLight {
		return lightChartPalette()
	}
	return darkChartPalette()
}

func darkChartPalette() chartPalette {
	return chartPalette{
		background:      color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		wheel:           color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		subtle:          color.NRGBA{R: 100, G: 110, B: 120, A: 255},
		sign:            color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		planet:          color.NRGBA{R: 38, G: 44, B: 52, A: 255},
		sun:             color.NRGBA{R: 184, G: 115, B: 33, A: 255},
		moon:            color.NRGBA{R: 74, G: 93, B: 116, A: 255},
		mercury:         color.NRGBA{R: 42, G: 108, B: 149, A: 255},
		venus:           color.NRGBA{R: 48, G: 128, B: 80, A: 255},
		mars:            color.NRGBA{R: 178, G: 55, B: 48, A: 255},
		jupiter:         color.NRGBA{R: 153, G: 104, B: 35, A: 255},
		saturn:          color.NRGBA{R: 87, G: 82, B: 73, A: 255},
		house:           color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		houseNumber:     color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		tick:            color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		accent:          color.NRGBA{R: 165, G: 60, B: 45, A: 255},
		secondaryAccent: color.NRGBA{R: 42, G: 96, B: 138, A: 255},
		easyAspect:      color.NRGBA{R: 42, G: 128, B: 87, A: 190},
		hardAspect:      color.NRGBA{R: 186, G: 55, B: 51, A: 190},
		neutralAspect:   color.NRGBA{R: 110, G: 116, B: 125, A: 130},
		text:            color.NRGBA{R: 40, G: 46, B: 55, A: 255},
		mutedText:       color.NRGBA{R: 99, G: 107, B: 119, A: 240},
	}
}

func lightChartPalette() chartPalette {
	return chartPalette{
		background:      color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		wheel:           color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		subtle:          color.NRGBA{R: 100, G: 110, B: 120, A: 255},
		sign:            color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		planet:          color.NRGBA{R: 38, G: 44, B: 52, A: 255},
		sun:             color.NRGBA{R: 184, G: 115, B: 33, A: 255},
		moon:            color.NRGBA{R: 74, G: 93, B: 116, A: 255},
		mercury:         color.NRGBA{R: 42, G: 108, B: 149, A: 255},
		venus:           color.NRGBA{R: 48, G: 128, B: 80, A: 255},
		mars:            color.NRGBA{R: 178, G: 55, B: 48, A: 255},
		jupiter:         color.NRGBA{R: 153, G: 104, B: 35, A: 255},
		saturn:          color.NRGBA{R: 87, G: 82, B: 73, A: 255},
		house:           color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		houseNumber:     color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		tick:            color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		accent:          color.NRGBA{R: 165, G: 60, B: 45, A: 255},
		secondaryAccent: color.NRGBA{R: 42, G: 96, B: 138, A: 255},
		easyAspect:      color.NRGBA{R: 42, G: 128, B: 87, A: 190},
		hardAspect:      color.NRGBA{R: 186, G: 55, B: 51, A: 190},
		neutralAspect:   color.NRGBA{R: 110, G: 116, B: 125, A: 130},
		text:            color.NRGBA{R: 40, G: 46, B: 55, A: 255},
		mutedText:       color.NRGBA{R: 99, G: 107, B: 119, A: 240},
	}
}

func inCanvasChartInfo(chart astro.Chart, palette chartPalette, size fyne.Size) []fyne.CanvasObject {
	x := 12.0
	y := 18.0
	headerSize := float32(12)
	bodySize := float32(10)

	subtitle := "Event Chart"
	if chart.ChartType != "" {
		subtitle = chart.ChartType.String() + " Chart"
	}

	objects := []fyne.CanvasObject{
		textAt(chart.Name, palette.text, headerSize+2, x, y, true, assets.CourierFont),
	}
	y += 18
	objects = append(objects, textAt(subtitle, palette.text, bodySize, x, y, true, assets.CourierFont))
	y += 16
	objects, _ = appendChartInfoDetails(objects, chart, palette, x, y, bodySize)

	return objects
}

func appendChartInfoDetails(objects []fyne.CanvasObject, chart astro.Chart, palette chartPalette, x, y float64, bodySize float32) ([]fyne.CanvasObject, float64) {
	localTime, zoneAbbr, formattedOffset := chartDisplayTime(chart)

	objects = append(objects, textAt(localTime.Format("2 Jan 2006, Mon"), palette.mutedText, bodySize, x, y, false, assets.CourierFont))
	y += 14
	objects = append(objects, textAt(fmt.Sprintf("%s %s %s", localTime.Format("15:04:05"), zoneAbbr, formattedOffset), palette.mutedText, bodySize, x, y, false, assets.CourierFont))
	y += 14
	objects = append(objects, textAt(shortenLocationName(chart.LocationName), palette.text, bodySize, x, y, false, assets.CourierFont))
	y += 14
	objects = append(objects, textAt(formatCoordsDMS(chart.Latitude, chart.Longitude), palette.text, bodySize, x, y, false, assets.CourierFont))
	y += 18

	objects = append(objects, textAtItalic("Geocentric", palette.mutedText, bodySize-1, x, y, true, assets.CourierFont))
	y += 12
	objects = append(objects, textAtItalic("Tropical", palette.mutedText, bodySize-1, x, y, true, assets.CourierFont))
	y += 12
	objects = append(objects, textAtItalic(chart.HouseSystem.Label(), palette.mutedText, bodySize-1, x, y, true, assets.CourierFont))
	y += 12
	objects = append(objects, textAtItalic("Mean Node", palette.mutedText, bodySize-1, x, y, true, assets.CourierFont))
	y += 12

	return objects, y
}

func chartDisplayTime(chart astro.Chart) (time.Time, string, string) {
	localTime := chart.DateTimeUTC
	zoneAbbr := "UTC"
	formattedOffset := "+00:00"

	if chart.TimezoneName != "" {
		loc, err := time.LoadLocation(chart.TimezoneName)
		if err == nil {
			localTime = chart.DateTimeUTC.In(loc)
			var offsetSec int
			zoneAbbr, offsetSec = localTime.Zone()
			offsetHours := float64(offsetSec) / 3600.0
			sign := "+"
			if offsetHours < 0 {
				sign = "-"
				offsetHours = -offsetHours
			}
			hours := int(offsetHours)
			minutes := int((offsetHours - float64(hours)) * 60)
			formattedOffset = fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
		}
	} else if chart.UTCOffset != "" {
		offsetHours, err := strconv.ParseFloat(chart.UTCOffset, 64)
		if err == nil {
			offsetSec := int(offsetHours * 3600)
			loc := time.FixedZone("local", offsetSec)
			localTime = chart.DateTimeUTC.In(loc)
			zoneAbbr = "LMT"
			sign := "+"
			if offsetHours < 0 {
				sign = "-"
				offsetHours = -offsetHours
			}
			hours := int(offsetHours)
			minutes := int((offsetHours - float64(hours)) * 60)
			formattedOffset = fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
		}
	}
	return localTime, zoneAbbr, formattedOffset
}

func panelBlock(x, y, width, height float64, palette chartPalette) fyne.CanvasObject {
	rect := canvas.NewRectangle(withAlpha(palette.background, 228))
	rect.StrokeColor = withAlpha(palette.subtle, 150)
	rect.StrokeWidth = 1
	rect.Resize(fyne.NewSize(float32(width), float32(height)))
	rect.Move(fyne.NewPos(float32(x), float32(y)))
	return rect
}

func textAt(value string, clr color.Color, size float32, x, y float64, bold bool, font fyne.Resource) fyne.CanvasObject {
	text := canvas.NewText(value, clr)
	text.TextSize = size
	text.TextStyle = fyne.TextStyle{Bold: bold}
	text.FontSource = font
	text.Move(fyne.NewPos(float32(x), float32(y)))
	return text
}

func textAtItalic(value string, clr color.Color, size float32, x, y float64, italic bool, font fyne.Resource) fyne.CanvasObject {
	text := canvas.NewText(value, clr)
	text.TextSize = size
	text.TextStyle = fyne.TextStyle{Italic: italic}
	text.FontSource = font
	text.Move(fyne.NewPos(float32(x), float32(y)))
	return text
}

func angularMarkerLabelPosition(label string, x, y float64, textSize float32) fyne.Position {
	width := float32(len(label)) * textSize * 0.62
	height := textSize * 0.55
	switch label {
	case "ASC":
		return fyne.NewPos(float32(x)-width-textSize*0.08, float32(y)-height)
	case "DSC":
		return fyne.NewPos(float32(x)+textSize*0.12, float32(y)-height)
	case "MC":
		return fyne.NewPos(float32(x)-width*0.5, float32(y)-textSize*1.15)
	case "IC":
		return fyne.NewPos(float32(x)-width*0.5, float32(y)+textSize*0.2)
	default:
		return fyne.NewPos(float32(x)-width*0.5, float32(y)-height)
	}
}

func houseCuspLongitude(chart astro.Chart, houseNumber int) float64 {
	for _, house := range chart.Houses {
		if house.Number == houseNumber {
			return house.CuspLongitude
		}
	}
	return 0
}

func houseLabelLongitude(houses []astro.House, index int) float64 {
	if len(houses) == 0 {
		return 0
	}
	current := houses[index].CuspLongitude
	next := houses[(index+1)%len(houses)].CuspLongitude
	if next < current {
		next += 360
	}
	return astro.NormalizeDegrees(current + (next-current)/2)
}

func withAlpha(clr color.Color, alpha uint8) color.Color {
	r, g, b, _ := clr.RGBA()
	return color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: alpha}
}

func shortPlanetName(planet astro.Planet) string {
	switch planet {
	case astro.Sun:
		return "Sun"
	case astro.Moon:
		return "Moon"
	case astro.Mercury:
		return "Merc"
	case astro.Venus:
		return "Ven"
	case astro.Mars:
		return "Mars"
	case astro.Jupiter:
		return "Jup"
	case astro.Saturn:
		return "Sat"
	default:
		if spec, ok := astro.ChartObjectSpecFor(planet); ok {
			return spec.Name
		}
		return string(planet)
	}
}

func background(size fyne.Size, fill color.Color) fyne.CanvasObject {
	rect := canvas.NewRectangle(fill)
	rect.Resize(size)
	rect.Move(fyne.NewPos(0, 0))
	return rect
}

func circle(cx, cy, radius float64, stroke color.Color, width float32) fyne.CanvasObject {
	c := canvas.NewCircle(color.Transparent)
	c.StrokeColor = stroke
	c.StrokeWidth = width
	c.Resize(fyne.NewSize(float32(radius*2), float32(radius*2)))
	c.Move(fyne.NewPos(float32(cx-radius), float32(cy-radius)))
	return c
}

func filledCircle(cx, cy, radius float64, fill, stroke color.Color, width float32) fyne.CanvasObject {
	c := canvas.NewCircle(fill)
	c.StrokeColor = stroke
	c.StrokeWidth = width
	c.Resize(fyne.NewSize(float32(radius*2), float32(radius*2)))
	c.Move(fyne.NewPos(float32(cx-radius), float32(cy-radius)))
	return c
}

func line(x1, y1, x2, y2 float64, stroke color.Color, width float32) fyne.CanvasObject {
	l := canvas.NewLine(stroke)
	l.StrokeWidth = width
	l.Position1 = fyne.NewPos(float32(x1), float32(y1))
	l.Position2 = fyne.NewPos(float32(x2), float32(y2))
	return l
}

func planetLabelObjects(position astro.PlanetPosition, centerX, centerY, x, y float64, planetTextSize, detailTextSize float32, clr color.Color) []fyne.CanvasObject {
	detailSize := detailTextSize * 0.82
	objects := make([]fyne.CanvasObject, 0, 5)

	glyph := canvas.NewText(planetGlyph(position.Planet), clr)
	glyph.TextSize = planetTextSize
	glyph.FontSource = astrologyFont()
	moveTextCentered(glyph, x, y)
	objects = append(objects, glyph)

	degrees, minutes := zodiacDegreeMinuteParts(position.Longitude)
	inwardX, inwardY := inwardUnit(centerX, centerY, x, y)
	degreeX := x + inwardX*float64(planetTextSize)*0.68
	degreeY := y + inwardY*float64(planetTextSize)*0.68
	signX := x + inwardX*float64(planetTextSize)*1.14
	signY := y + inwardY*float64(planetTextSize)*1.14
	minuteX := x + inwardX*float64(planetTextSize)*1.6
	minuteY := y + inwardY*float64(planetTextSize)*1.6
	degreeText := canvas.NewText(fmt.Sprintf("%02d", degrees), color.Black)
	degreeText.TextSize = detailSize
	degreeText.FontSource = assets.CourierFont
	degreeText.Move(centeredTextPosition(degreeX, degreeY, detailSize))
	objects = append(objects, degreeText)

	signText := canvas.NewText(signGlyph(astro.SignFromLongitude(position.Longitude)), clr)
	signText.TextSize = detailSize * 1.05
	signText.FontSource = astrologyFont()
	moveTextCentered(signText, signX, signY)
	objects = append(objects, signText)

	minuteText := canvas.NewText(fmt.Sprintf("%02d", minutes), color.Black)
	minuteText.TextSize = detailSize
	minuteText.FontSource = assets.CourierFont
	minuteText.Move(centeredTextPosition(minuteX, minuteY, detailSize))
	objects = append(objects, minuteText)

	if position.Retrograde {
		retrograde := canvas.NewText("R", clr)
		retrograde.TextSize = detailSize * 0.78
		retrograde.FontSource = assets.CourierFont
		retrograde.Move(fyne.NewPos(float32(minuteX)+detailSize*0.52, float32(minuteY)-detailSize*0.42))
		objects = append(objects, retrograde)
	}
	return objects
}

func centeredTextPosition(x, y float64, textSize float32) fyne.Position {
	return fyne.NewPos(float32(x)-textSize*0.58, float32(y)-textSize*0.5)
}

func moveTextCentered(text *canvas.Text, x, y float64) {
	size := text.MinSize()
	text.Move(fyne.NewPos(float32(x)-size.Width/2, float32(y)-size.Height/2))
}

func planetLineEndpoint(centerX, centerY, x, y, outwardOffset float64) (float64, float64) {
	inwardX, inwardY := inwardUnit(centerX, centerY, x, y)
	return x - inwardX*outwardOffset, y - inwardY*outwardOffset
}

func inwardUnit(centerX, centerY, x, y float64) (float64, float64) {
	dx := centerX - x
	dy := centerY - y
	length := math.Hypot(dx, dy)
	if length == 0 {
		return 0, 0
	}
	return dx / length, dy / length
}

func planetClusterGuide(x, y float64, textSize float32, clr color.Color) fyne.CanvasObject {
	return line(
		x-float64(textSize)*0.42,
		y+float64(textSize)*0.42,
		x+float64(textSize)*0.85,
		y+float64(textSize)*0.42,
		withAlpha(clr, 145),
		0.65,
	)
}

func zodiacDegreeMinuteParts(longitude float64) (int, int) {
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

func point(cx, cy, radius, longitude float64) (float64, float64) {
	angle := (longitude - 90) * math.Pi / 180
	return cx + math.Cos(angle)*radius, cy + math.Sin(angle)*radius
}

func chartPoint(cx, cy, radius, longitude, ascendant float64) (float64, float64) {
	displayAngle := 180 + astro.NormalizeDegrees(longitude-ascendant)
	radian := displayAngle * math.Pi / 180
	return cx + math.Cos(radian)*radius, cy - math.Sin(radian)*radius
}

type planetPlacementCandidate struct {
	x float64
	y float64
}

type planetPlacementObstacle struct {
	x     float64
	y     float64
	lineX float64
	lineY float64
}

func planetPlacements(planets []astro.PlanetPosition, centerX, centerY, anchorRadius, baseRadius, ascendant, step, labelSize float64) []planetPlacement {
	sorted := append([]astro.PlanetPosition(nil), planets...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Longitude < sorted[j].Longitude
	})

	placements := make([]planetPlacement, 0, len(sorted))
	for _, cluster := range closePlanetClusters(sorted, 4.5) {
		placements = append(placements, optimizedPlanetCluster(cluster, sorted, centerX, centerY, anchorRadius, baseRadius, ascendant, step, labelSize)...)
	}
	return placements
}

func optimizedPlanetCluster(cluster, allPlanets []astro.PlanetPosition, centerX, centerY, anchorRadius, baseRadius, ascendant, step, labelSize float64) []planetPlacement {
	if len(cluster) == 1 {
		x, y := chartPoint(centerX, centerY, baseRadius, cluster[0].Longitude, ascendant)
		return []planetPlacement{{position: cluster[0], x: x, y: y}}
	}
	if len(cluster) > 5 {
		return fallbackPlanetCluster(cluster, centerX, centerY, baseRadius, ascendant, step)
	}

	candidates := make([][]planetPlacementCandidate, len(cluster))
	for i, position := range cluster {
		candidates[i] = planetPlacementCandidates(centerX, centerY, baseRadius, position.Longitude, ascendant, step)
	}
	obstacles := planetPlacementObstacles(allPlanets, cluster, centerX, centerY, anchorRadius, baseRadius, ascendant)

	best := make([]planetPlacementCandidate, len(cluster))
	current := make([]planetPlacementCandidate, len(cluster))
	bestScore := math.Inf(1)
	var search func(int)
	search = func(index int) {
		if index == len(cluster) {
			score := planetPlacementScore(cluster, current, obstacles, centerX, centerY, anchorRadius, baseRadius, ascendant, labelSize)
			if score < bestScore {
				bestScore = score
				copy(best, current)
			}
			return
		}
		for _, candidate := range candidates[index] {
			current[index] = candidate
			search(index + 1)
		}
	}
	search(0)

	placements := make([]planetPlacement, 0, len(cluster))
	for i, position := range cluster {
		placements = append(placements, planetPlacement{
			position: position,
			x:        best[i].x,
			y:        best[i].y,
		})
	}
	return placements
}

func planetPlacementObstacles(allPlanets, cluster []astro.PlanetPosition, centerX, centerY, anchorRadius, baseRadius, ascendant float64) []planetPlacementObstacle {
	obstacles := make([]planetPlacementObstacle, 0, len(allPlanets)-len(cluster))
	for _, position := range allPlanets {
		if clusterContainsPlanet(cluster, position) {
			continue
		}
		x, y := chartPoint(centerX, centerY, baseRadius, position.Longitude, ascendant)
		lineX, lineY := chartPoint(centerX, centerY, anchorRadius, position.Longitude, ascendant)
		obstacles = append(obstacles, planetPlacementObstacle{
			x:     x,
			y:     y,
			lineX: lineX,
			lineY: lineY,
		})
	}
	return obstacles
}

func clusterContainsPlanet(cluster []astro.PlanetPosition, position astro.PlanetPosition) bool {
	for _, member := range cluster {
		if member.Planet == position.Planet {
			return true
		}
	}
	return false
}

func fallbackPlanetCluster(cluster []astro.PlanetPosition, centerX, centerY, baseRadius, ascendant, step float64) []planetPlacement {
	placements := make([]planetPlacement, 0, len(cluster))
	center := float64(len(cluster)-1) / 2
	for i, position := range cluster {
		x, y := shiftedPoint(centerX, centerY, baseRadius, position.Longitude, ascendant, (float64(i)-center)*step, 0)
		placements = append(placements, planetPlacement{position: position, x: x, y: y})
	}
	return placements
}

func planetPlacementCandidates(centerX, centerY, baseRadius, longitude, ascendant, step float64) []planetPlacementCandidate {
	tangentOffsets := []float64{-step * 0.75, -step * 0.35, 0, step * 0.35, step * 0.75}
	candidates := make([]planetPlacementCandidate, 0, len(tangentOffsets))
	for _, tangentOffset := range tangentOffsets {
		x, y := shiftedPoint(centerX, centerY, baseRadius, longitude, ascendant, tangentOffset, 0)
		candidates = append(candidates, planetPlacementCandidate{x: x, y: y})
	}
	return candidates
}

func shiftedPoint(cx, cy, radius, longitude, ascendant, tangentOffset, radialOffset float64) (float64, float64) {
	x, y := chartPoint(cx, cy, radius, longitude, ascendant)
	dx := x - cx
	dy := y - cy
	length := math.Hypot(dx, dy)
	if length == 0 {
		return x, y
	}

	radialX := dx / length
	radialY := dy / length
	tangentX := -radialY
	tangentY := radialX
	return x + tangentX*tangentOffset + radialX*radialOffset, y + tangentY*tangentOffset + radialY*radialOffset
}

func planetPlacementScore(cluster []astro.PlanetPosition, candidates []planetPlacementCandidate, obstacles []planetPlacementObstacle, centerX, centerY, anchorRadius, baseRadius, ascendant, labelSize float64) float64 {
	score := 0.0
	for i, candidate := range candidates {
		baseX, baseY := chartPoint(centerX, centerY, anchorRadius, cluster[i].Longitude, ascendant)
		naturalX, naturalY := chartPoint(centerX, centerY, baseRadius, cluster[i].Longitude, ascendant)
		score += math.Hypot(candidate.x-naturalX, candidate.y-naturalY) * 1.4
		for _, obstacle := range obstacles {
			score += labelOverlapPenalty(candidate, planetPlacementCandidate{x: obstacle.x, y: obstacle.y}, labelSize) * 1.8
			if segmentsIntersect(baseX, baseY, candidate.x, candidate.y, obstacle.lineX, obstacle.lineY, obstacle.x, obstacle.y) {
				score += 260
			}
		}
		for j := i + 1; j < len(candidates); j++ {
			score += labelOverlapPenalty(candidate, candidates[j], labelSize)
			nextBaseX, nextBaseY := chartPoint(centerX, centerY, anchorRadius, cluster[j].Longitude, ascendant)
			if segmentsIntersect(baseX, baseY, candidate.x, candidate.y, nextBaseX, nextBaseY, candidates[j].x, candidates[j].y) {
				score += 420
			}
		}
	}
	return score
}

func labelOverlapPenalty(a, b planetPlacementCandidate, labelSize float64) float64 {
	width := labelSize * 1.55
	height := labelSize * 2.8
	overlapX := width - math.Abs(a.x-b.x)
	overlapY := height - math.Abs(a.y-b.y)
	if overlapX <= 0 || overlapY <= 0 {
		return 0
	}
	return overlapX*overlapY*8 + 500
}

func segmentsIntersect(ax, ay, bx, by, cx, cy, dx, dy float64) bool {
	return orientation(ax, ay, bx, by, cx, cy)*orientation(ax, ay, bx, by, dx, dy) < 0 &&
		orientation(cx, cy, dx, dy, ax, ay)*orientation(cx, cy, dx, dy, bx, by) < 0
}

func orientation(ax, ay, bx, by, cx, cy float64) float64 {
	return (bx-ax)*(cy-ay) - (by-ay)*(cx-ax)
}

func closePlanetClusters(planets []astro.PlanetPosition, maxGap float64) [][]astro.PlanetPosition {
	if len(planets) == 0 {
		return nil
	}

	clusters := [][]astro.PlanetPosition{{planets[0]}}
	for _, position := range planets[1:] {
		current := clusters[len(clusters)-1]
		previous := current[len(current)-1]
		if angularGap(previous.Longitude, position.Longitude) <= maxGap {
			clusters[len(clusters)-1] = append(current, position)
			continue
		}
		clusters = append(clusters, []astro.PlanetPosition{position})
	}

	if len(clusters) > 1 {
		first := clusters[0]
		last := clusters[len(clusters)-1]
		if angularGap(last[len(last)-1].Longitude, first[0].Longitude) <= maxGap {
			merged := append(append([]astro.PlanetPosition{}, last...), first...)
			clusters = append([][]astro.PlanetPosition{merged}, clusters[1:len(clusters)-1]...)
		}
	}
	return clusters
}

func angularGap(a, b float64) float64 {
	distance := math.Abs(astro.NormalizeDegrees(a) - astro.NormalizeDegrees(b))
	if distance > 180 {
		return 360 - distance
	}
	return distance
}

func clamp(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func aspectColor(typ astro.AspectType, palette chartPalette) color.Color {
	switch typ {
	case astro.Trine, astro.Sextile:
		return palette.easyAspect
	case astro.Square, astro.Opposition:
		return palette.hardAspect
	default:
		return palette.neutralAspect
	}
}

func planetColor(planet astro.Planet, palette chartPalette) color.Color {
	switch planet {
	case astro.Sun:
		return palette.sun
	case astro.Moon:
		return palette.moon
	case astro.Mercury:
		return palette.mercury
	case astro.Venus:
		return palette.venus
	case astro.Mars:
		return palette.mars
	case astro.Jupiter:
		return palette.jupiter
	case astro.Saturn:
		return palette.saturn
	default:
		return palette.planet
	}
}

func planetLongitude(chart astro.Chart, planet astro.Planet) (float64, bool) {
	for _, position := range chart.Planets {
		if position.Planet == planet {
			return position.Longitude, true
		}
	}
	return 0, false
}
