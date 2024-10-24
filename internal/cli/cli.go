package cli

import (
	"flag"
	"fmt"

	"github.com/moneyforward-i/admina-sysutils/internal/logger"
)

// Run executes the CLI application with the given arguments
func Run(args []string) error {
	flags := flag.NewFlagSet("admina-sysutils", flag.ExitOnError)
	helpFlag := flags.Bool("help", false, "Show help")
	debugMode := flags.Bool("debug", false, "Enable debug mode")

	if err := flags.Parse(args); err != nil {
		return err
	}

	logger.Init(*debugMode)

	if *helpFlag || len(args) == 0 {
		printHelp()
		return nil
	}

	switch flags.Arg(0) {
	case "identity":
		cmd := NewIdentityCommand()
		return cmd.Run(flags.Args()[1:])
	default:
		return fmt.Errorf("unknown command: %s\nRun 'admina-sysutils --help' for usage", flags.Arg(0))
	}
}

// printHelp displays the help message for the main command
func printHelp() {
	fmt.Print(`Usage: admina-sysutils [--help] [--debug] <command> [subcommand]

Options:
  --help     Show help
  --debug    Enable debug mode

Commands:
  identity   Identity management commands`)
}
