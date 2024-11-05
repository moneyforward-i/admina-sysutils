package identity_test

import (
	"testing"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	"github.com/moneyforward-i/admina-sysutils/internal/identity"
	mock "github.com/moneyforward-i/admina-sysutils/internal/identity/testing"
	"github.com/moneyforward-i/admina-sysutils/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestGetIdentityMatrix(t *testing.T) {
	mockClient := &mock.MockClient{
		Identities: []admina.Identity{
			{ID: "1", ManagementType: "internal", EmployeeStatus: "active", Email: "test1@example.com"},
			{ID: "2", ManagementType: "external", EmployeeStatus: "inactive", Email: "test2@example.com"},
		},
	}

	matrix, err := identity.GetIdentityMatrix(mockClient)
	assert.NoError(t, err)
	assert.NotNil(t, matrix)
	assert.Equal(t, 2, len(matrix.ManagementTypes))
	assert.Equal(t, 2, len(matrix.Statuses))
}

func TestPrintIdentityMatrix(t *testing.T) {
	// Loggerの初期化
	logger.Init()

	mockClient := &mock.MockClient{
		Identities: []admina.Identity{
			{ID: "1", ManagementType: "internal", EmployeeStatus: "active", Email: "test1@example.com"},
			{ID: "2", ManagementType: "external", EmployeeStatus: "inactive", Email: "test2@example.com"},
		},
	}

	err := identity.PrintIdentityMatrix(mockClient, "json")
	assert.NoError(t, err)
}
