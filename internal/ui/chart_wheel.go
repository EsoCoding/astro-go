package ui

import (
	"fmt"
	"image/color"
	"math"

	"astro-go/internal/astro"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

const chartWheelSize float32 = 560

func NewChartWheel(chart astro.Chart) fyne.CanvasObject {
	center := float64(chartWheelSize / 2)
	outer := float64(chartWheelSize * 0.46)
	inner := float64(chartWheelSize * 0.34)
	planetRadius := float64(chartWheelSize * 0.285)
	signRadius := float64(chartWheelSize * 0.41)

	objects := []fyne.CanvasObject{
		circle(center, center, outer, color.NRGBA{R: 42, G: 45, B: 52, A: 255}, 2),
		circle(center, center, inner, color.NRGBA{R: 88, G: 95, B: 108, A: 255}, 1),
	}

	for i := 0; i < 12; i++ {
		longitude := float64(i * 30)
		x1, y1 := point(center, center, inner, longitude)
		x2, y2 := point(center, center, outer, longitude)
		objects = append(objects, line(x1, y1, x2, y2, color.NRGBA{R: 86, G: 92, B: 104, A: 255}, 1))

		labelLong := longitude + 15
		x, y := point(center, center, signRadius, labelLong)
		text := canvas.NewText(astro.Sign(i).Glyph(), color.NRGBA{R: 34, G: 38, B: 46, A: 255})
		text.TextSize = 13
		text.TextStyle.Bold = true
		text.Move(fyne.NewPos(float32(x-10), float32(y-8)))
		objects = append(objects, text)
	}

	for _, house := range chart.Houses {
		longitude := float64(int(chart.Ascendant.Sign)+house.Number-1) * 30
		x1, y1 := point(center, center, inner*0.72, longitude)
		x2, y2 := point(center, center, outer, longitude)
		objects = append(objects, line(x1, y1, x2, y2, color.NRGBA{R: 126, G: 82, B: 70, A: 255}, 1))
	}

	for _, aspect := range chart.Aspects {
		from, okFrom := planetLongitude(chart, aspect.From)
		to, okTo := planetLongitude(chart, aspect.To)
		if !okFrom || !okTo {
			continue
		}
		x1, y1 := point(center, center, inner*0.65, from)
		x2, y2 := point(center, center, inner*0.65, to)
		objects = append(objects, line(x1, y1, x2, y2, aspectColor(aspect.Type), 1))
	}

	for i, planet := range chart.Planets {
		x, y := point(center, center, planetRadius-float64(i%2)*18, planet.Longitude)
		text := canvas.NewText(planet.Planet.Glyph(), color.NRGBA{R: 15, G: 23, B: 42, A: 255})
		text.TextSize = 12
		text.TextStyle.Bold = true
		text.Move(fyne.NewPos(float32(x-10), float32(y-8)))
		objects = append(objects, text)
	}

	ascText := canvas.NewText(fmt.Sprintf("Asc %s", formatZodiac(chart.Ascendant.Longitude)), color.NRGBA{R: 126, G: 42, B: 40, A: 255})
	ascText.TextStyle.Bold = true
	ascText.TextSize = 13
	ascText.Move(fyne.NewPos(18, chartWheelSize-34))
	objects = append(objects, ascText)

	mcText := canvas.NewText(fmt.Sprintf("MC %s", formatZodiac(chart.MC.Longitude)), color.NRGBA{R: 42, G: 82, B: 126, A: 255})
	mcText.TextStyle.Bold = true
	mcText.TextSize = 13
	mcText.Move(fyne.NewPos(18, chartWheelSize-54))
	objects = append(objects, mcText)

	wheel := container.NewWithoutLayout(objects...)
	wheel.Resize(fyne.NewSize(chartWheelSize, chartWheelSize))
	return wheel
}

func circle(cx, cy, radius float64, stroke color.Color, width float32) fyne.CanvasObject {
	c := canvas.NewCircle(color.Transparent)
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

func point(cx, cy, radius, longitude float64) (float64, float64) {
	angle := (longitude - 90) * math.Pi / 180
	return cx + math.Cos(angle)*radius, cy + math.Sin(angle)*radius
}

func aspectColor(typ astro.AspectType) color.Color {
	switch typ {
	case astro.Trine, astro.Sextile:
		return color.NRGBA{R: 46, G: 112, B: 82, A: 170}
	case astro.Square, astro.Opposition:
		return color.NRGBA{R: 154, G: 56, B: 52, A: 170}
	default:
		return color.NRGBA{R: 87, G: 90, B: 98, A: 120}
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
