package cli

import (
	"flag"
	"fmt"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	"github.com/moneyforward-i/admina-sysutils/internal/identity"
)

// IdentityCommand handles all identity-related subcommands
type IdentityCommand struct {
	flags        *flag.FlagSet
	outputFormat *string
	debugMode    *bool
}

// NewIdentityCommand creates and initializes a new IdentityCommand
func NewIdentityCommand() *IdentityCommand {
	cmd := &IdentityCommand{
		flags: flag.NewFlagSet("identity", flag.ExitOnError),
	}
	cmd.outputFormat = cmd.flags.String("output", "json", "Output format (json, markdown, pretty)")
	cmd.debugMode = cmd.flags.Bool("debug", false, "Enable debug mode")
	return cmd
}

// Run executes the identity command with the given arguments
func (c *IdentityCommand) Run(args []string) error {
	if err := c.flags.Parse(args); err != nil {
		return err
	}

	if len(c.flags.Args()) == 0 {
		return fmt.Errorf("subcommand is required")
	}

	switch c.flags.Arg(0) {
	case "matrix":
		return c.runMatrix()
	default:
		return fmt.Errorf("unknown subcommand: %s", c.flags.Arg(0))
	}
}

// Help returns the help message for the identity command
func (c *IdentityCommand) Help() string {
	return `Usage: admina-sysutils identity [options] <subcommand>

Subcommands:
  matrix    Display identity matrix

Options:
  --output format    Output format (json, markdown, pretty)
  --debug           Enable debug mode
`
}

// runMatrix executes the matrix subcommand
func (c *IdentityCommand) runMatrix() error {
	client := c.newIdentityClient()
	if client == nil {
		return fmt.Errorf("failed to initialize client")
	}

	return identity.PrintIdentityMatrix(client, *c.outputFormat)
}

// newIdentityClient creates a new identity client with the current debug mode
func (c *IdentityCommand) newIdentityClient() identity.Client {
	client := admina.NewClient(*c.debugMode)
	if client == nil {
		return nil
	}

	if err := client.Validate(); err != nil {
		return nil
	}

	return &identityClientAdapter{client: client}
}

// identityClientAdapter adapts admina.Client to identity.Client interface
type identityClientAdapter struct {
	client *admina.Client
}

// GetIdentities retrieves identities from the Admina API and converts them to the internal format
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
