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

func TestProxyURLEncoding(t *testing.T) {
	tests := []struct {
		name        string
		inputURL    string
		expected    string
		shouldError bool
	}{
		{
			name:     "backslash in username",
			inputURL: "http://SSC\\999999:Sakatase99@172.16.10.10:2335",
			expected: "http://SSC%5C999999:Sakatase99@172.16.10.10:2335",
		},
		{
			name:     "special characters in password",
			inputURL: "http://user:pass@word@proxy.example.com:8080",
			expected: "http://user:pass%2540word@proxy.example.com:8080",
		},
		{
			name:     "multiple special characters",
			inputURL: "http://domain\\user@name:p@ss:w0rd@proxy.local:3128",
			expected: "http://domain%5Cuser%40name:p%40ss:w0rd@proxy.local:3128",
		},
		{
			name:     "no special characters",
			inputURL: "http://normaluser:normalpass@proxy.test.com:8080",
			expected: "http://normaluser:normalpass@proxy.test.com:8080",
		},
		{
			name:     "proxy without user credentials",
			inputURL: "http://proxy.example.com:8080",
			expected: "http://proxy.example.com:8080",
		},
		{
			name:     "https proxy without user credentials",
			inputURL: "https://proxy.example.com:8080",
			expected: "https://proxy.example.com:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 元の環境変数を保存
			originalHTTPProxy := os.Getenv("HTTP_PROXY")
			originalHTTPSProxy := os.Getenv("HTTPS_PROXY")
			originalOrgID := os.Getenv("ADMINA_ORGANIZATION_ID")
			originalAPIKey := os.Getenv("ADMINA_API_KEY")

			// 環境変数を設定
			os.Setenv("HTTP_PROXY", tt.inputURL)
			os.Unsetenv("HTTPS_PROXY")
			os.Setenv("ADMINA_ORGANIZATION_ID", "test-org")
			os.Setenv("ADMINA_API_KEY", "test-key")

			// NewClientを実行（プロキシURLのエンコーディングが実行される）
			_ = NewClient()

			// 環境変数から実際の結果を取得
			encodedProxy := os.Getenv("HTTP_PROXY")

			if encodedProxy != tt.expected {
				t.Errorf("Expected encoded URL: %s, got: %s", tt.expected, encodedProxy)
			}

			// 環境変数を復元
			if originalHTTPProxy != "" {
				os.Setenv("HTTP_PROXY", originalHTTPProxy)
			} else {
				os.Unsetenv("HTTP_PROXY")
			}
			if originalHTTPSProxy != "" {
				os.Setenv("HTTPS_PROXY", originalHTTPSProxy)
			} else {
				os.Unsetenv("HTTPS_PROXY")
			}
			if originalOrgID != "" {
				os.Setenv("ADMINA_ORGANIZATION_ID", originalOrgID)
			} else {
				os.Unsetenv("ADMINA_ORGANIZATION_ID")
			}
			if originalAPIKey != "" {
				os.Setenv("ADMINA_API_KEY", originalAPIKey)
			} else {
				os.Unsetenv("ADMINA_API_KEY")
			}
		})
	}
}

func TestNewClientWithSpecialCharacterProxy(t *testing.T) {
	// 元の環境変数を保存
	originalHTTPProxy := os.Getenv("HTTP_PROXY")
	originalHTTPSProxy := os.Getenv("HTTPS_PROXY")
	originalOrgID := os.Getenv("ADMINA_ORGANIZATION_ID")
	originalAPIKey := os.Getenv("ADMINA_API_KEY")

	// テスト用環境変数を設定
	os.Setenv("HTTP_PROXY", "http://SSC\\999999:Sakatase99@172.16.10.10:2335")
	os.Unsetenv("HTTPS_PROXY")
	os.Setenv("ADMINA_ORGANIZATION_ID", "test-org")
	os.Setenv("ADMINA_API_KEY", "test-key")

	// クライアント作成（エラーが発生しないことを確認）
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	// 環境変数がエンコードされていることを確認
	encodedProxy := os.Getenv("HTTP_PROXY")
	expected := "http://SSC%5C999999:Sakatase99@172.16.10.10:2335"
	if encodedProxy != expected {
		t.Errorf("Expected encoded proxy URL in environment: %s, got: %s", expected, encodedProxy)
	}

	// 環境変数を復元
	if originalHTTPProxy != "" {
		os.Setenv("HTTP_PROXY", originalHTTPProxy)
	} else {
		os.Unsetenv("HTTP_PROXY")
	}
	if originalHTTPSProxy != "" {
		os.Setenv("HTTPS_PROXY", originalHTTPSProxy)
	} else {
		os.Unsetenv("HTTPS_PROXY")
	}
	if originalOrgID != "" {
		os.Setenv("ADMINA_ORGANIZATION_ID", originalOrgID)
	} else {
		os.Unsetenv("ADMINA_ORGANIZATION_ID")
	}
	if originalAPIKey != "" {
		os.Setenv("ADMINA_API_KEY", originalAPIKey)
	} else {
		os.Unsetenv("ADMINA_API_KEY")
	}
}

func TestProxyComponentValidation(t *testing.T) {
	tests := []struct {
		name           string
		proxyURL       string
		envSchema      string
		envUser        string
		envPassword    string
		envHost        string
		expectMismatch bool
		mismatchType   string
	}{
		{
			name:           "all components match",
			proxyURL:       "http://SSC\\999999:Sakatase99@172.16.10.10:2335",
			envSchema:      "http",
			envUser:        "SSC\\999999",
			envPassword:    "Sakatase99",
			envHost:        "172.16.10.10:2335",
			expectMismatch: false,
		},
		{
			name:           "schema mismatch",
			proxyURL:       "http://user:pass@host:8080",
			envSchema:      "https",
			envUser:        "user",
			envPassword:    "pass",
			envHost:        "host:8080",
			expectMismatch: true,
			mismatchType:   "schema",
		},
		{
			name:           "user mismatch",
			proxyURL:       "http://user1:pass@host:8080",
			envSchema:      "http",
			envUser:        "user2",
			envPassword:    "pass",
			envHost:        "host:8080",
			expectMismatch: true,
			mismatchType:   "user",
		},
		{
			name:           "password mismatch",
			proxyURL:       "http://user:pass1@host:8080",
			envSchema:      "http",
			envUser:        "user",
			envPassword:    "pass2",
			envHost:        "host:8080",
			expectMismatch: true,
			mismatchType:   "password",
		},
		{
			name:           "host mismatch",
			proxyURL:       "http://user:pass@host1:8080",
			envSchema:      "http",
			envUser:        "user",
			envPassword:    "pass",
			envHost:        "host2:8080",
			expectMismatch: true,
			mismatchType:   "host",
		},
		{
			name:           "no environment variables set",
			proxyURL:       "http://user:pass@host:8080",
			envSchema:      "",
			envUser:        "",
			envPassword:    "",
			envHost:        "",
			expectMismatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 元の環境変数を保存
			originalHTTPProxy := os.Getenv("HTTP_PROXY")
			originalHTTPSProxy := os.Getenv("HTTPS_PROXY")
			originalOrgID := os.Getenv("ADMINA_ORGANIZATION_ID")
			originalAPIKey := os.Getenv("ADMINA_API_KEY")
			originalProxySchema := os.Getenv("PROXY_SCHEMA")
			originalProxyUser := os.Getenv("PROXY_USER")
			originalProxyPassword := os.Getenv("PROXY_PASSWORD")
			originalProxyHost := os.Getenv("PROXY_HOST")

			// テスト用環境変数を設定
			os.Setenv("HTTP_PROXY", tt.proxyURL)
			os.Unsetenv("HTTPS_PROXY")
			os.Setenv("ADMINA_ORGANIZATION_ID", "test-org")
			os.Setenv("ADMINA_API_KEY", "test-key")

			// プロキシコンポーネント検証用の環境変数を設定
			if tt.envSchema != "" {
				os.Setenv("PROXY_SCHEMA", tt.envSchema)
			} else {
				os.Unsetenv("PROXY_SCHEMA")
			}
			if tt.envUser != "" {
				os.Setenv("PROXY_USER", tt.envUser)
			} else {
				os.Unsetenv("PROXY_USER")
			}
			if tt.envPassword != "" {
				os.Setenv("PROXY_PASSWORD", tt.envPassword)
			} else {
				os.Unsetenv("PROXY_PASSWORD")
			}
			if tt.envHost != "" {
				os.Setenv("PROXY_HOST", tt.envHost)
			} else {
				os.Unsetenv("PROXY_HOST")
			}

			// NewClientを実行（プロキシURLの検証が実行される）
			_ = NewClient()

			// 環境変数を復元
			if originalHTTPProxy != "" {
				os.Setenv("HTTP_PROXY", originalHTTPProxy)
			} else {
				os.Unsetenv("HTTP_PROXY")
			}
			if originalHTTPSProxy != "" {
				os.Setenv("HTTPS_PROXY", originalHTTPSProxy)
			} else {
				os.Unsetenv("HTTPS_PROXY")
			}
			if originalOrgID != "" {
				os.Setenv("ADMINA_ORGANIZATION_ID", originalOrgID)
			} else {
				os.Unsetenv("ADMINA_ORGANIZATION_ID")
			}
			if originalAPIKey != "" {
				os.Setenv("ADMINA_API_KEY", originalAPIKey)
			} else {
				os.Unsetenv("ADMINA_API_KEY")
			}
			if originalProxySchema != "" {
				os.Setenv("PROXY_SCHEMA", originalProxySchema)
			} else {
				os.Unsetenv("PROXY_SCHEMA")
			}
			if originalProxyUser != "" {
				os.Setenv("PROXY_USER", originalProxyUser)
			} else {
				os.Unsetenv("PROXY_USER")
			}
			if originalProxyPassword != "" {
				os.Setenv("PROXY_PASSWORD", originalProxyPassword)
			} else {
				os.Unsetenv("PROXY_PASSWORD")
			}
			if originalProxyHost != "" {
				os.Setenv("PROXY_HOST", originalProxyHost)
			} else {
				os.Unsetenv("PROXY_HOST")
			}
		})
	}
}
