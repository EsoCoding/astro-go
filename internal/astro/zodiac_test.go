package astro

import "testing"

func TestSignFromLongitude(t *testing.T) {
	tests := []struct {
		longitude float64
		want      Sign
	}{
		{0, Aries},
		{29.99, Aries},
		{30, Taurus},
		{359.99, Pisces},
		{-1, Pisces},
	}

	for _, test := range tests {
		if got := SignFromLongitude(test.longitude); got != test.want {
			t.Fatalf("SignFromLongitude(%f) = %s, want %s", test.longitude, got, test.want)
		}
	}
}

func TestWholeSignHouse(t *testing.T) {
	if got := WholeSignHouse(Aries, 75); got != 3 {
		t.Fatalf("WholeSignHouse() = %d, want 3", got)
	}
	if got := WholeSignHouse(Scorpio, 10); got != 6 {
		t.Fatalf("WholeSignHouse() = %d, want 6", got)
	}
}

func TestEssentialStatus(t *testing.T) {
	if got := EssentialStatus(Mars, Aries); got != "domicile" {
		t.Fatalf("EssentialStatus() = %s, want domicile", got)
	}
	if got := EssentialStatus(Sun, Libra); got != "fall" {
		t.Fatalf("EssentialStatus() = %s, want fall", got)
	}
}

func TestHamburgSymbolGlyphs(t *testing.T) {
	if Aries.Glyph() != "a" {
		t.Fatalf("Aries glyph = %q, want HamburgSymbols key a", Aries.Glyph())
	}
	if Sun.Glyph() != "Q" {
		t.Fatalf("Sun glyph = %q, want HamburgSymbols key Q", Sun.Glyph())
	}
	if NorthNode.Glyph() != "\u008b" {
		t.Fatalf("North Node glyph = %q, want HamburgSymbols key \\u008b", NorthNode.Glyph())
	}
	if SouthNode.Glyph() != "\u008c" {
		t.Fatalf("South Node glyph = %q, want HamburgSymbols key \\u008c", SouthNode.Glyph())
	}
	if Chiron.Glyph() != "M" {
		t.Fatalf("Chiron glyph = %q, want HamburgSymbols key M", Chiron.Glyph())
	}
	if ParsFortunae.Glyph() != "<" {
		t.Fatalf("Pars Fortunae glyph = %q, want HamburgSymbols key <", ParsFortunae.Glyph())
	}
}
