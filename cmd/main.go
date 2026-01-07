package main

import (
	"fmt"
	"os"

	"github.com/andro-kes/userflux/internal/orchestrator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: loadtest <script> [args...]")
		os.Exit(1)
	}

	script := os.Args[1]

	if err := orchestrator.Orchestrator(script); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
