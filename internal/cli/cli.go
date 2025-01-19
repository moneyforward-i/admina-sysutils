package cli

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	"github.com/moneyforward-i/admina-sysutils/internal/logger"
	"github.com/moneyforward-i/admina-sysutils/internal/organization"
)

// Run executes the CLI application with the given arguments
func Run(args []string) error {
	flags := flag.NewFlagSet("admina-sysutils", flag.ExitOnError)
	helpFlag := flags.Bool("help", false, "Show help")
	debugFlag := flags.Bool("debug", false, "Enable debug mode")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *debugFlag {
		os.Setenv("ADMINA_DEBUG", "true")
	}
	logger.Init()

	if *helpFlag || len(args) == 0 {
		printHelp()
		return nil
	}

	client := admina.NewClient()
	if client == nil {
		return fmt.Errorf("failed to initialize client")
	}

	ctx := context.Background()
	org, err := client.GetOrganization(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization info: %v", err)
	}

	organization.PrintInfo(org)

	if err := executeCommand(flags, client); err != nil {
		return err
	}
	organization.PrintInfo(org)
	return nil
}

// executeCommand handles subcommand execution
func executeCommand(flags *flag.FlagSet, client *admina.Client) error {
	switch flags.Arg(0) {
	case "identity":
		cmd := NewIdentityCommand()
		return cmd.Run(flags.Args()[1:])
	default:
		return fmt.Errorf("unknown command: %s\nRun 'admina-sysutils --help' for usage", flags.Arg(0))
	}
}

func printHelp() {
	logger.Print(`Usage: admina-sysutils [--help] [--debug] <command> [subcommand]

Options:
  --help     Show help
  --debug    Enable debug mode

Commands:
  identity   Identity management commands`)
}
