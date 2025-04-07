package admina

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/moneyforward-i/admina-sysutils/internal/logger"
)

func init() {
	logger.Init()
}

func setupTestServer() (*httptest.Server, *Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/organizations/test-org":
			json.NewEncoder(w).Encode(Organization{
				ID:   1,
				Name: "Test Org",
			})
		case "/api/v1/organizations/test-org/identity":
			response := APIResponse[[]Identity]{
				Meta: Meta{
					NextCursor: "next",
				},
				Items: []Identity{
					{
						ID:          "1",
						PeopleID:    1,
						DisplayName: "Test User",
						Email:       "test@example.com",
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		case "/api/v1/organizations/test-org/identity/merge":
			// Revert to using Raw JSON String for the mock response
			rawJSONResponse := `{
				"meta": {
					"next_cursor": "next"
				},
				"items": [
					{
						"id": "1",
						"peopleId": 2,
						"displayName": "Test User",
						"primaryEmail": "test@example.com",
						"secondaryEmails": [],
						"managementType": "managed",
						"employeeType": "full_time_employee",
						"employeeStatus": "active",
						"mergedPeople": [
							{
								"id": 1,
								"displayName": "",
								"primaryEmail": "from@example.com",
								"username": ""
							}
						]
					}
				],
				"dummy_feature_field": "dummy_feature_value"
			}`
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(rawJSONResponse))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	client := NewClient()
	client.baseURL = server.URL + "/api/v1"
	client.organizationID = "test-org"
	client.apiKey = "test-key"

	return server, client
}

func TestNewClient(t *testing.T) {
	// 環境変数の設定
	originalBaseURL := os.Getenv("ADMINA_BASE_URL")
	originalOrgID := os.Getenv("ADMINA_ORGANIZATION_ID")
	originalAPIKey := os.Getenv("ADMINA_API_KEY")
	defer func() {
		os.Setenv("ADMINA_BASE_URL", originalBaseURL)
		os.Setenv("ADMINA_ORGANIZATION_ID", originalOrgID)
		os.Setenv("ADMINA_API_KEY", originalAPIKey)
	}()

	tests := []struct {
		name        string
		setupEnv    func()
		wantBaseURL string
	}{
		{
			name: "default configuration",
			setupEnv: func() {
				os.Unsetenv("ADMINA_BASE_URL")
				os.Setenv("ADMINA_ORGANIZATION_ID", "test-org")
				os.Setenv("ADMINA_API_KEY", "test-key")
			},
			wantBaseURL: DefaultBaseURL,
		},
		{
			name: "custom base URL",
			setupEnv: func() {
				os.Setenv("ADMINA_BASE_URL", "https://custom.example.com")
				os.Setenv("ADMINA_ORGANIZATION_ID", "test-org")
				os.Setenv("ADMINA_API_KEY", "test-key")
			},
			wantBaseURL: "https://custom.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			client := NewClient()
			if client.baseURL != tt.wantBaseURL {
				t.Errorf("NewClient() baseURL = %v, want %v", client.baseURL, tt.wantBaseURL)
			}
		})
	}
}

func TestGetIdentities(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()

	ctx := context.Background()
	identities, nextCursor, err := client.GetIdentities(ctx, "")
	if err != nil {
		t.Fatalf("GetIdentities() error = %v", err)
	}

	if len(identities) != 1 {
		t.Errorf("GetIdentities() got %d identities, want 1", len(identities))
	}

	if nextCursor != "next" {
		t.Errorf("GetIdentities() nextCursor = %v, want 'next'", nextCursor)
	}
}

func TestMergeIdentities(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()

	ctx := context.Background()
	result, err := client.MergeIdentities(ctx, 1, 2)
	if err != nil {
		t.Fatalf("MergeIdentities() error = %v", err)
	}

	// 戻り値の検証
	if result.FromPeopleID != 1 || result.ToPeopleID != 2 {
		t.Errorf("MergeIdentities() got = %v, want FromPeopleID=1, ToPeopleID=2", result)
	}
}

func TestGetOrganization(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()

	ctx := context.Background()
	org, err := client.GetOrganization(ctx)
	if err != nil {
		t.Fatalf("GetOrganization() error = %v", err)
	}

	if org.ID != 1 || org.Name != "Test Org" {
		t.Errorf("GetOrganization() returned unexpected organization")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		orgID   string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "valid configuration",
			orgID:   "test-org",
			apiKey:  "test-key",
			wantErr: false,
		},
		{
			name:    "missing organization ID",
			orgID:   "",
			apiKey:  "test-key",
			wantErr: true,
		},
		{
			name:    "missing API key",
			orgID:   "test-org",
			apiKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				organizationID: tt.orgID,
				apiKey:         tt.apiKey,
			}
			if err := client.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAPIError_Error(t *testing.T) {
	apiErr := &APIError{
		StatusCode: 404,
		Message:    "Not Found",
		Body:       "Resource not found",
		RequestID:  "req-123",
		Timestamp:  time.Now(),
	}

	errStr := apiErr.Error()
	if errStr == "" {
		t.Error("APIError.Error() returned empty string")
	}
}
