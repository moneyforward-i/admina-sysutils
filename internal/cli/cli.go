package cli

import (
	"flag"
	"fmt"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	"github.com/moneyforward-i/admina-sysutils/internal/identity"
	"github.com/moneyforward-i/admina-sysutils/internal/logger"
)

var (
	debugMode    bool
	outputFormat string
)

func Run(args []string) error {
	flags := flag.NewFlagSet("admina-sysutils", flag.ExitOnError)
	helpFlag := flags.Bool("help", false, "Show help")
	flags.BoolVar(&debugMode, "debug", false, "Enable debug mode")
	flags.StringVar(&outputFormat, "output", "json", "Specify output format (json, markdown, pretty)")

	if err := flags.Parse(args); err != nil {
		return err
	}

	logger.Init(debugMode)

	if *helpFlag || len(args) == 0 {
		printHelp()
		return nil
	}

	switch flags.Arg(0) {
	case "identity":
		return runIdentityCommand(flags.Args()[1:])
	default:
		return fmt.Errorf("unknown command: %s\nRun 'admina-sysutils --help' for usage", flags.Arg(0))
	}
}

func runIdentityCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("identity subcommand requires arguments")
	}

	switch args[0] {
	case "matrix":
		return printIdentityMatrix()
	default:
		return fmt.Errorf("unknown identity subcommand: %s\nRun 'admina-sysutils --help' for usage", args[0])
	}
}

func printHelp() {
	fmt.Println("Usage: admina-sysutils [--help] [--debug] [--output format] <command> [subcommand]")
	fmt.Println("Options:")
	fmt.Println("  --help                   Show help")
	fmt.Println("  --debug                  Enable debug mode")
	fmt.Println("  --output format          Specify output format (json, markdown, pretty)")
	fmt.Println("\nCommands:")
	fmt.Println("  identity matrix          Display identity matrix")
}

func printIdentityMatrix() error {
	client := admina.NewClient(debugMode)
	if client == nil {
		return fmt.Errorf("failed to initialize client")
	}

	if err := client.Validate(); err != nil {
		return err
	}

	// Adapter to use admina.Client as identity.IdentityClient
	identityClient := &identityClientAdapter{client: client}

	return identity.PrintIdentityMatrix(identityClient, outputFormat)
}

// identityClientAdapter is an adapter to make admina.Client compatible with identity.IdentityClient interface
type identityClientAdapter struct {
	client *admina.Client
}

func (a *identityClientAdapter) GetIdentities(cursor string) ([]identity.Identity, string, int, error) {
	adminaIdentities, nextCursor, totalCount, err := a.client.GetIdentities(cursor)
	if err != nil {
		return nil, "", 0, err
	}

	identities := make([]identity.Identity, len(adminaIdentities))
	for i, ai := range adminaIdentities {
		identities[i] = identity.Identity{
			ManagementType: ai.ManagementType,
			EmployeeStatus: ai.EmployeeStatus,
			Status:         ai.Status,
		}
	}

	return identities, nextCursor, totalCount, nil
}
