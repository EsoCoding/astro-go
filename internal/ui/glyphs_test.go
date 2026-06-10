package ui

import (
	"testing"

	"astro-go/internal/astro"
)

func TestEnigmaPlanetGlyphs(t *testing.T) {
	tests := []struct {
		planet astro.Planet
		want   string
	}{
		{astro.Sun, "\uE200"},
		{astro.NorthNode, "\uE525"},
		{astro.SouthNode, "\uE526"},
		{astro.ParsFortunae, "\uF400"},
	}

	for _, test := range tests {
		if got := planetGlyph(test.planet); got != test.want {
			t.Fatalf("planetGlyph(%s) = %q, want %q", test.planet, got, test.want)
		}
	}
}

func TestEnigmaPlanetGlyphUnknownFallback(t *testing.T) {
	if got := planetGlyph(astro.Planet("UnknownInternalName")); got != "?" {
		t.Fatalf("unknown planet glyph fallback = %q, want ?", got)
	}
}
