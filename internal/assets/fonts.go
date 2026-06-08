package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed fonts/HamburgSymbols.ttf
var hamburgSymbols []byte

//go:embed fonts/courier.ttf
var courier []byte

var (
	HamburgSymbolsFont = fyne.NewStaticResource("HamburgSymbols.ttf", hamburgSymbols)
	CourierFont        = fyne.NewStaticResource("courier.ttf", courier)
)
