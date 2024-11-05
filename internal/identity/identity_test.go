package identity_test

import (
	"testing"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	mock "github.com/moneyforward-i/admina-sysutils/internal/admina/mock"
	"github.com/moneyforward-i/admina-sysutils/internal/identity"
	"github.com/stretchr/testify/assert"
)

func TestFetchAllIdentities(t *testing.T) {
	mockClient := &mock.Client{
		Identities: []admina.Identity{
			{ID: "1", ManagementType: "internal", EmployeeStatus: "active", Email: "test1@example.com"},
			{ID: "2", ManagementType: "external", EmployeeStatus: "inactive", Email: "test2@example.com"},
		},
		Cursor: "",
	}

	identities, err := identity.FetchAllIdentities(mockClient)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(identities))
	assert.Equal(t, "1", identities[0].ID)
	assert.Equal(t, "2", identities[1].ID)
}

func TestMaskEmail(t *testing.T) {
	email := "example@domain.com"
	maskedEmail := identity.MaskEmail(email)
	assert.Equal(t, "ex*****@domain.com", maskedEmail)

	invalidEmail := "invalid-email"
	maskedInvalidEmail := identity.MaskEmail(invalidEmail)
	assert.Equal(t, "invalid-email", maskedInvalidEmail)
}

func TestExtractDomain(t *testing.T) {
	email := "example@domain.com"
	domain := identity.ExtractDomain(email)
	assert.Equal(t, "domain.com", domain)

	invalidEmail := "invalid-email"
	invalidDomain := identity.ExtractDomain(invalidEmail)
	assert.Equal(t, "", invalidDomain)
}

func TestExtractLocalPart(t *testing.T) {
	email := "example@domain.com"
	localPart := identity.ExtractLocalPart(email)
	assert.Equal(t, "example", localPart)

	invalidEmail := "invalid-email"
	invalidLocalPart := identity.ExtractLocalPart(invalidEmail)
	assert.Equal(t, "", invalidLocalPart)
}
