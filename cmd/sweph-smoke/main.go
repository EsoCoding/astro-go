package main

import (
	"fmt"
	"os"

	"astro-go/internal/sweph"
)

func main() {
	result, err := sweph.Smoke()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Swiss Ephemeris version: %s\n", result.Version)
	fmt.Printf("Julian day: %.5f\n", result.JulianDay)
	fmt.Printf("Sun longitude: %.6f\n", result.SunLongitude)
	fmt.Printf("Sun latitude: %.6f\n", result.SunLatitude)
}
