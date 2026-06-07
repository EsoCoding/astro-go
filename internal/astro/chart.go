package astro

import "time"

type BirthData struct {
	Name             string
	DateTimeUTC      time.Time
	LatitudeDegrees  float64
	LongitudeDegrees float64
}

type PlanetPosition struct {
	Planet          Planet
	Longitude       float64
	Latitude        float64
	Speed           float64
	Sign            Sign
	DegreeInSign    float64
	WholeSignHouse  int
	Retrograde      bool
	DomicileRuler   Planet
	EssentialStatus string
}

type Angle struct {
	Name         string
	Longitude    float64
	Sign         Sign
	DegreeInSign float64
}

type House struct {
	Number int
	Sign   Sign
	Ruler  Planet
}

type Aspect struct {
	From  Planet
	To    Planet
	Type  AspectType
	Orb   float64
	Exact float64
}

type Chart struct {
	Name        string
	DateTimeUTC time.Time
	Latitude    float64
	Longitude   float64
	JulianDay   float64
	Ascendant   Angle
	MC          Angle
	Houses      []House
	Planets     []PlanetPosition
	Aspects     []Aspect
}
