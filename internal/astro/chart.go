package astro

import "time"

type BirthData struct {
	Name             string
	DateTimeUTC      time.Time
	LocationName     string
	LatitudeDegrees  float64
	LongitudeDegrees float64
	HouseSystem      HouseSystem
	UTCOffset        string
	TimezoneName     string
	ChartType        ChartType
	EnabledObjects   []Planet
}

type PlanetPosition struct {
	Planet          Planet
	Longitude       float64
	Latitude        float64
	Speed           float64
	Sign            Sign
	DegreeInSign    float64
	House           int
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
	Number        int
	CuspLongitude float64
	Sign          Sign
	Ruler         Planet
}

type Aspect struct {
	From  Planet
	To    Planet
	Type  AspectType
	Orb   float64
	Exact float64
}

type Chart struct {
	Name         string
	DateTimeUTC  time.Time
	LocationName string
	Latitude     float64
	Longitude    float64
	HouseSystem  HouseSystem
	JulianDay    float64
	Ascendant    Angle
	MC           Angle
	Houses       []House
	Planets      []PlanetPosition
	Aspects      []Aspect
	UTCOffset    string
	TimezoneName string
	ChartType    ChartType
}
