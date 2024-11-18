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
	// テスト開始時にマスク処理を有効化
	identity.SetNoMask(false)

	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "通常のメールアドレス",
			email:    "example@domain.com",
			expected: "exa****@domain.com",
		},
		{
			name:     "短いローカルパート",
			email:    "abc@domain.com",
			expected: "abc@domain.com",
		},
		{
			name:     "不正なメールアドレス",
			email:    "invalid-email",
			expected: "invalid-email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := identity.MaskEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
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
