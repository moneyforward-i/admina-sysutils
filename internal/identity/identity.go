package identity

import (
	"fmt"
	"os"
	"strings"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	"golang.org/x/net/context"
)

// Client interface defines the methods required for identity operations
type Client interface {
	GetIdentities(ctx context.Context, cursor string) ([]admina.Identity, string, error)
	MergeIdentities(ctx context.Context, fromPeopleID, toPeopleID int) error
}

// Common utility functions
func FetchAllIdentities(client Client) ([]admina.Identity, error) {
	var allIdentities []admina.Identity
	nextCursor := ""
	step := 0
	totalProcessed := 0

	for {
		step++
		fmt.Fprintf(os.Stderr, "\rProcessing step: %d (Total: %d)", step, totalProcessed)

		identities, cursor, err := client.GetIdentities(context.Background(), nextCursor)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n")
			return nil, fmt.Errorf("failed to fetch identities: %v", err)
		}

		allIdentities = append(allIdentities, identities...)
		totalProcessed += len(identities)

		if cursor == "" {
			break
		}
		nextCursor = cursor
	}

	fmt.Fprintf(os.Stderr, "\nProcessing complete. Total steps: %d\n", step)
	fmt.Fprintf(os.Stderr, "Number of Identities retrieved: %d\n", totalProcessed)

	return allIdentities, nil
}

// Email utility functions
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "invalid-email"
	}
	localPart := parts[0]
	if len(localPart) <= 2 {
		return "*****@" + parts[1]
	}
	return localPart[:2] + strings.Repeat("*", len(localPart)-2) + "@" + parts[1]
}

func ExtractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

func ExtractLocalPart(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}
