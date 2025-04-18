package identity

import (
	"fmt"
	"strings"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	"github.com/moneyforward-i/admina-sysutils/internal/logger"
	"golang.org/x/net/context"
)

// Client interface defines the methods required for identity operations
type Client interface {
	GetIdentities(ctx context.Context, cursor string) ([]admina.Identity, string, error)
	MergeIdentities(ctx context.Context, fromPeopleID, toPeopleID int) (admina.MergeIdentity, error)
}

// Common utility functions
func FetchAllIdentities(client Client) ([]admina.Identity, error) {
	var allIdentities []admina.Identity
	nextCursor := ""
	step := 0
	totalProcessed := 0

	for {
		step++
		logger.PrintErr("\rProcessing step: %d (Total: %d)", step, totalProcessed)

		identities, cursor, err := client.GetIdentities(context.Background(), nextCursor)
		if err != nil {
			logger.PrintErr("\n")
			return nil, fmt.Errorf("failed to fetch identities: %v", err)
		}

		allIdentities = append(allIdentities, identities...)

		totalProcessed += len(identities)

		if cursor == "" {
			break
		}
		nextCursor = cursor
	}

	logger.PrintErr("\nProcessing complete. Total steps: %d\n", step)
	logger.PrintErr("Number of Identities retrieved: %d\n", totalProcessed)

	return allIdentities, nil
}

var noMask bool

// SetNoMask はメールマスクの設定を行います
func SetNoMask(flag bool) {
	noMask = flag
}

// MaskEmail はメールアドレスをマスクします
func MaskEmail(email string) string {
	if noMask {
		return email
	}

	if !strings.Contains(email, "@") {
		return email
	}

	localPart, domain := ExtractLocalPart(email), ExtractDomain(email)
	if len(localPart) <= 2 {
		return email
	}

	return localPart[:3] + strings.Repeat("*", len(localPart)-3) + "@" + domain
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
