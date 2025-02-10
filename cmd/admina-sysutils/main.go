package main

import (
	"os"
	"strings"
	"time"

	"github.com/moneyforward-i/admina-sysutils/internal/cli"
	"github.com/moneyforward-i/admina-sysutils/internal/logger"
)

func main() {
	startTime := time.Now()

	logger.PrintErr("Executed command: > %s\n", strings.TrimPrefix(strings.Join(os.Args, " "), os.Args[0]+" "))

	err := cli.Run(os.Args[1:])

	duration := time.Since(startTime)
	logger.PrintErr("Processing time: %v\n", duration)

	if err != nil {
		logger.PrintErr("Error: %v\n", err)
		logger.PrintErr("Result: 1 (Error)\n")
		os.Exit(1)
	}

	logger.PrintErr("Result: 0 (Success)\n")
}
