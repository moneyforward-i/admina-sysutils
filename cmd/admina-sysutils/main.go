package main

import (
	"fmt"
	"os"

	"github.com/yourusername/admina-sysutils/internal/cli"
)

func main() {
	if err := cli.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
}
