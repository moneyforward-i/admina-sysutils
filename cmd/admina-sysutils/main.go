package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/moneyforward-i/admina-sysutils/internal/cli"
)

// main function
func main() {
	startTime := time.Now()

	fmt.Fprintf(os.Stderr, "Executed command: > %s\n", strings.TrimPrefix(strings.Join(os.Args, " "), os.Args[0]+" "))

	err := cli.Run(os.Args[1:])

	duration := time.Since(startTime)
	fmt.Fprintf(os.Stderr, "Processing time: %v\n", duration)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Result: 1 (Error)\n")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Result: 0 (Success)\n")
}
