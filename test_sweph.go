package main

import (
	"fmt"

	swisseph "github.com/tejzpr/go-swisseph"
)

func main() {
	fmt.Printf("Uranus: %d\n", swisseph.Uranus)
	fmt.Printf("Neptune: %d\n", swisseph.Neptune)
	fmt.Printf("Pluto: %d\n", swisseph.Pluto)
	fmt.Printf("MeanNode: %d\n", swisseph.MeanNode)
	fmt.Printf("TrueNode: %d\n", swisseph.TrueNode)
	fmt.Printf("Chiron: %d\n", swisseph.Chiron)
}
