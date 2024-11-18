package identity_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	mock "github.com/moneyforward-i/admina-sysutils/internal/admina/mock"
	"github.com/moneyforward-i/admina-sysutils/internal/identity"
	"github.com/moneyforward-i/admina-sysutils/internal/logger"
	"github.com/stretchr/testify/assert"
)

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
				Identities: []admina.Identity{
					{ID: "1", ManagementType: "internal", EmployeeStatus: "active", Email: "test1@parent-domain.com"},
					{ID: "2", ManagementType: "external", EmployeeStatus: "active", Email: "test1@child-domain.com"},
					{ID: "3", ManagementType: "external", EmployeeStatus: "active", Email: "unmapped@child-domain.com"},
				},
			}

			config := &identity.MergeConfig{
				ParentDomain: "parent-domain.com",
				ChildDomains: []string{"child-domain.com"},
				DryRun:       true,
				AutoApprove:  true,
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
							assert.Contains(t, string(mappingsContent), "test1@parent-domain.com")
							assert.Contains(t, string(mappingsContent), "test1@child-domain.com")
						} else {
							assert.Contains(t, string(mappingsContent), "tes**@parent-domain.com")
							assert.Contains(t, string(mappingsContent), "tes**@child-domain.com")
						}

						unmappedContent, err := os.ReadFile(unmappedPath)
						assert.NoError(t, err)
						if tc.noMask {
							assert.Contains(t, string(unmappedContent), "unmapped@child-domain.com")
						} else {
							assert.Contains(t, string(unmappedContent), "unm*****@child-domain.com")
						}
					}
				})
			}
		})
	}
}
