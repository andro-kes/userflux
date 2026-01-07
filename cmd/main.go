package main

import (
	"fmt"
	"os"

	"github.com/andro-kes/userflux/internal/logging"
	"github.com/andro-kes/userflux/internal/orchestrator"
)

func main() {
	// Initialize logger
	logger, err := logging.NewLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("userflux CLI starting")

	if len(os.Args) < 2 {
		logger.Error("No script provided")
		fmt.Fprintln(os.Stderr, "usage: loadtest <script> [args...]")
		os.Exit(1)
	}

	script := os.Args[1]
	logger.Infof("Selected script: %s", script)

	if err := orchestrator.Orchestrator(script, logger); err != nil {
		logger.Errorf("Orchestrator error: %v", err)
		fmt.Println(err.Error())
		os.Exit(1)
	}

	logger.Info("userflux CLI completed successfully")
}
