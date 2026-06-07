package main

import (
	"fmt"
	"os"

	"astro-go/internal/app"
	"astro-go/internal/ui"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) > 0 && (args[0] == "-v" || args[0] == "--version") {
		fmt.Println(app.Version())
		return nil
	}

	ui.Launch()
	return nil
}
