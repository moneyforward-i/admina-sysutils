package identity_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	mock "github.com/moneyforward-i/admina-sysutils/internal/admina/mock"
	"github.com/moneyforward-i/admina-sysutils/internal/identity"
	"github.com/moneyforward-i/admina-sysutils/internal/logger"
	"github.com/stretchr/testify/assert"
)

// テストで使用する共通のIdentityデータ
var testIdentities = []admina.Identity{
	{ID: "100", PeopleID: 101, ManagementType: "managed", EmployeeStatus: "active", Email: "user1@parent.domain.com"},
	{ID: "200", PeopleID: 202, ManagementType: "external", EmployeeStatus: "active", Email: "user1@child.domain.com"},
	{ID: "300", ManagementType: "external", EmployeeStatus: "active", Email: "unmapped@child.domain.com"},
}

func TestMergeIdentitiesWithFormatters(t *testing.T) {
	// Loggerの初期化
	logger.Init()

	// テストケースごとにNoMaskの設定を変えてテスト
	testCases := []struct {
		name   string
		noMask bool
	}{
		{"WithMask", false},
		{"WithoutMask", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// NoMaskの設定
			identity.SetNoMask(tc.noMask)

			mockClient := &mock.Client{
				Identities: testIdentities,
			}

			config := &identity.MergeConfig{
				ParentDomain: "parent.domain.com",
				ChildDomains: []string{"child.domain.com"},
				DryRun:       true,
				AutoApprove:  true,
				OutputFormat: "json",
			}

			projectRoot, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get working directory: %v", err)
			}
			projectRoot = filepath.Dir(filepath.Dir(projectRoot))
			os.Setenv("ADMINA_CLI_ROOT", projectRoot)

			formats := []string{"json", "markdown", "pretty", "csv"}
			for _, format := range formats {
				t.Run(format, func(t *testing.T) {
					config.OutputFormat = format
					err := identity.MergeIdentities(mockClient, config)
					assert.NoError(t, err)

					if format == "csv" {
						// CSVファイルの出力ディレクトリを確認
						outputDir := filepath.Join(projectRoot, "out", "data")
						logger.LogInfo("Output directory: %s", outputDir)
						mappingsPath := filepath.Join(outputDir, "identity_mappings.csv")
						unmappedPath := filepath.Join(outputDir, "unmapped_child_identities.csv")

						// CSVファイルの存在確認
						assert.FileExists(t, mappingsPath)
						assert.FileExists(t, unmappedPath)

						// CSVファイルの内容を確認
						mappingsContent, err := os.ReadFile(mappingsPath)
						assert.NoError(t, err)
						if tc.noMask {
							assert.Contains(t, string(mappingsContent), "user1@parent.domain.com")
							assert.Contains(t, string(mappingsContent), "user1@child.domain.com")
						} else {
							assert.Contains(t, string(mappingsContent), "use**@parent.domain.com")
							assert.Contains(t, string(mappingsContent), "use**@child.domain.com")
						}

						unmappedContent, err := os.ReadFile(unmappedPath)
						assert.NoError(t, err)
						if tc.noMask {
							assert.Contains(t, string(unmappedContent), "unmapped@child.domain.com")
						} else {
							assert.Contains(t, string(unmappedContent), "unm*****@child.domain.com")
						}
					}
				})
			}
		})
	}
}

func TestMergeIdentitiesOrder(t *testing.T) {
	logger.Init()

	// テストデータの準備
	mockClient := &mock.Client{
		Identities: testIdentities,
	}

	config := &identity.MergeConfig{
		ParentDomain: "parent.domain.com",
		ChildDomains: []string{"child.domain.com"},
		DryRun:       false,
		AutoApprove:  true,
		OutputFormat: "json",
	}

	// テストの実行
	err := identity.MergeIdentities(mockClient, config)
	assert.NoError(t, err)

	// マージ結果の検証
	t.Run("Verify merge direction", func(t *testing.T) {
		assert.Len(t, mockClient.MergeResults, 1, "Should have exactly one merge result")
		assert.Equal(t, 202, mockClient.MergeResults[0].FromPeopleID, "FromPeopleID should be from child domain")
		assert.Equal(t, 101, mockClient.MergeResults[0].ToPeopleID, "ToPeopleID should be from parent domain")
	})
}

func TestMergeIdentitiesError(t *testing.T) {
	logger.Init()

	testCases := []struct {
		name          string
		mockClient    *mock.Client
		config        *identity.MergeConfig
		expectedError string
	}{
		{
			name: "マージが許可されていないケース",
			mockClient: &mock.Client{
				Identities: []admina.Identity{
					{ID: "100", PeopleID: 101, ManagementType: "external", EmployeeStatus: "active", Email: "user1@parent.domain.com"},
					{ID: "200", PeopleID: 202, ManagementType: "managed", EmployeeStatus: "active", Email: "user1@child.domain.com"},
				},
			},
			config: &identity.MergeConfig{
				ParentDomain: "parent.domain.com",
				ChildDomains: []string{"child.domain.com"},
				DryRun:       false,
				AutoApprove:  true,
				OutputFormat: "json",
			},
			expectedError: "",
		},
		{
			name: "マージ実行時のエラーケース",
			mockClient: &mock.Client{
				Identities: testIdentities,
				Error:      fmt.Errorf("merge failed"),
			},
			config: &identity.MergeConfig{
				ParentDomain: "parent.domain.com",
				ChildDomains: []string{"child.domain.com"},
				DryRun:       false,
				AutoApprove:  true,
				OutputFormat: "json",
			},
			expectedError: "failed to fetch identities: failed to fetch identities: merge failed",
		},
		{
			name: "不正なフォーマット指定のケース",
			mockClient: &mock.Client{
				Identities: testIdentities,
			},
			config: &identity.MergeConfig{
				ParentDomain: "parent.domain.com",
				ChildDomains: []string{"child.domain.com"},
				DryRun:       false,
				AutoApprove:  true,
				OutputFormat: "invalid",
			},
			expectedError: "unknown output format: invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := identity.MergeIdentities(tc.mockClient, tc.config)
			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMergeIdentitiesDryRun(t *testing.T) {
	logger.Init()

	mockClient := &mock.Client{
		Identities: testIdentities,
	}

	config := &identity.MergeConfig{
		ParentDomain: "parent.domain.com",
		ChildDomains: []string{"child.domain.com"},
		DryRun:       true,
		AutoApprove:  true,
		OutputFormat: "json",
	}

	err := identity.MergeIdentities(mockClient, config)
	assert.NoError(t, err)
	assert.Empty(t, mockClient.MergeResults, "ドライランモードではマージは実行されないはずです")
}

func TestMergeAllowedPatterns(t *testing.T) {
	logger.Init()

	testCases := []struct {
		name          string
		parentType    string
		childType     string
		expectAllowed bool
	}{
		// managed のパターン
		{"managed to managed", "managed", "managed", true},
		{"managed to external", "external", "managed", false},
		{"managed to system", "system", "managed", false},
		{"managed to unregistered", "unregistered", "managed", false},
		{"managed to unknown", "unknown", "managed", false},

		// external のパターン
		{"external to managed", "managed", "external", true},
		{"external to external", "external", "external", true},
		{"external to system", "system", "external", false},
		{"external to unregistered", "unregistered", "external", false},
		{"external to unknown", "unknown", "external", false},

		// system のパターン
		{"system to managed", "managed", "system", true},
		{"system to external", "external", "system", true},
		{"system to system", "system", "system", true},
		{"system to unregistered", "unregistered", "system", false},
		{"system to unknown", "unknown", "system", false},

		// unregistered のパターン
		{"unregistered to managed", "managed", "unregistered", true},
		{"unregistered to external", "external", "unregistered", true},
		{"unregistered to system", "system", "unregistered", true},
		{"unregistered to unregistered", "unregistered", "unregistered", false},
		{"unregistered to unknown", "unknown", "unregistered", false},

		// unknown のパターン
		{"unknown to managed", "managed", "unknown", true},
		{"unknown to external", "external", "unknown", true},
		{"unknown to system", "system", "unknown", true},
		{"unknown to unregistered", "unregistered", "unknown", true},
		{"unknown to unknown", "unknown", "unknown", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parent := admina.Identity{ManagementType: tc.parentType}
			child := admina.Identity{ManagementType: tc.childType}
			result := identity.IsMergeAllowed(parent, child)
			assert.Equal(t, tc.expectAllowed, result,
				"Unexpected result for %s to %s merge", tc.childType, tc.parentType)
		})
	}
}

func TestMergeIdentitiesAPIErrors(t *testing.T) {
	logger.Init()

	testCases := []struct {
		name          string
		setupMock     func() *mock.Client
		expectedError string
	}{
		{
			name: "APIタイムアウトエラー",
			setupMock: func() *mock.Client {
				return &mock.Client{
					Identities: testIdentities,
					Error:      context.DeadlineExceeded,
				}
			},
			expectedError: "failed to fetch identities: failed to fetch identities: context deadline exceeded",
		},
		{
			name: "ネットワークエラー",
			setupMock: func() *mock.Client {
				return &mock.Client{
					Identities: testIdentities,
					Error:      &net.OpError{Op: "dial", Net: "tcp", Err: fmt.Errorf("connection refused")},
				}
			},
			expectedError: "failed to fetch identities: failed to fetch identities: dial tcp: connection refused",
		},
		{
			name: "認証エラー",
			setupMock: func() *mock.Client {
				return &mock.Client{
					Identities: testIdentities,
					Error: &admina.APIError{
						StatusCode: http.StatusUnauthorized,
						Message:    "unauthorized",
					},
				}
			},
			expectedError: "failed to fetch identities: failed to fetch identities: API error: status=401",
		},
		{
			name: "不正なレスポンス形式",
			setupMock: func() *mock.Client {
				return &mock.Client{
					Identities: testIdentities,
					Error:      fmt.Errorf("failed to decode response: invalid character"),
				}
			},
			expectedError: "failed to fetch identities: failed to fetch identities: failed to decode response: invalid character",
		},
		{
			name: "マージ実行時のエラー",
			setupMock: func() *mock.Client {
				client := &mock.Client{
					Identities: testIdentities,
				}
				client.Error = fmt.Errorf("merge operation failed")
				return client
			},
			expectedError: "failed to fetch identities: merge operation failed",
		},
	}

	config := &identity.MergeConfig{
		ParentDomain: "parent.domain.com",
		ChildDomains: []string{"child.domain.com"},
		DryRun:       false,
		AutoApprove:  true,
		OutputFormat: "json",
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := tc.setupMock()
			err := identity.MergeIdentities(mockClient, config)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}
