package identity

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	"github.com/stretchr/testify/assert"
)

func TestFindMergeCandidates(t *testing.T) {
	testCases := []struct {
		name           string
		identities     []admina.Identity
		config         *MergeConfig
		expectedResult *MergeResult
		expectError    bool
	}{
		{
			name: "Basic matching case",
			config: &MergeConfig{
				ParentDomain: "parent-domain.com",
				ChildDomains: []string{"child-domain.com"},
			},
			identities: []admina.Identity{
				{
					ManagementType: "managed",
					EmployeeStatus: "active",
					Email:          "test@parent-domain.com",
				},
				{
					ManagementType: "unregistered",
					EmployeeStatus: "active",
					Email:          "test@child-domain.com",
				},
			},
			expectedResult: &MergeResult{
				Candidates: []MergeCandidate{
					{
						Parent: admina.Identity{
							ManagementType: "managed",
							EmployeeStatus: "active",
							Email:          "test@parent-domain.com",
						},
						Child: admina.Identity{
							ManagementType: "unregistered",
							EmployeeStatus: "active",
							Email:          "test@child-domain.com",
						},
					},
				},
				Unmapped: []admina.Identity{},
			},
			expectError: false,
		},
		{
			name: "No matches",
			config: &MergeConfig{
				ParentDomain: "parent-domain.com",
				ChildDomains: []string{"child-domain.com"},
			},
			identities: []admina.Identity{
				{
					ManagementType: "managed",
					EmployeeStatus: "active",
					Email:          "test1@parent-domain.com",
				},
				{
					ManagementType: "unregistered",
					EmployeeStatus: "active",
					Email:          "test2@child-domain.com",
				},
			},
			expectedResult: &MergeResult{
				Candidates: []MergeCandidate{},
				Unmapped: []admina.Identity{
					{
						ManagementType: "unregistered",
						EmployeeStatus: "active",
						Email:          "test2@child-domain.com",
					},
				},
			},
			expectError: false,
		},
		{
			name: "Multiple child domains",
			config: &MergeConfig{
				ParentDomain: "parent-domain.com",
				ChildDomains: []string{"child1-domain.com", "child2-domain.com"},
			},
			identities: []admina.Identity{
				{
					ManagementType: "managed",
					EmployeeStatus: "active",
					Email:          "test1@parent-domain.com",
				},
				{
					ManagementType: "managed",
					EmployeeStatus: "active",
					Email:          "test2@parent-domain.com",
				},
				{
					ManagementType: "unregistered",
					EmployeeStatus: "active",
					Email:          "test1@child1-domain.com",
				},
				{
					ManagementType: "unregistered",
					EmployeeStatus: "active",
					Email:          "test2@child1-domain.com",
				},
				{
					ManagementType: "unregistered",
					EmployeeStatus: "active",
					Email:          "test1@child2-domain.com",
				},
				{
					ManagementType: "unregistered",
					EmployeeStatus: "active",
					Email:          "unmapped@child2-domain.com",
				},
			},
			expectedResult: &MergeResult{
				Candidates: []MergeCandidate{
					{
						Parent: admina.Identity{
							ManagementType: "managed",
							EmployeeStatus: "active",
							Email:          "test1@parent-domain.com",
						},
						Child: admina.Identity{
							ManagementType: "unregistered",
							EmployeeStatus: "active",
							Email:          "test1@child1-domain.com",
						},
					},
					{
						Parent: admina.Identity{
							ManagementType: "managed",
							EmployeeStatus: "active",
							Email:          "test1@parent-domain.com",
						},
						Child: admina.Identity{
							ManagementType: "unregistered",
							EmployeeStatus: "active",
							Email:          "test1@child2-domain.com",
						},
					},
					{
						Parent: admina.Identity{
							ManagementType: "managed",
							EmployeeStatus: "active",
							Email:          "test2@parent-domain.com",
						},
						Child: admina.Identity{
							ManagementType: "unregistered",
							EmployeeStatus: "active",
							Email:          "test2@child1-domain.com",
						},
					},
				},
				Unmapped: []admina.Identity{
					{
						ManagementType: "unregistered",
						EmployeeStatus: "active",
						Email:          "unmapped@child2-domain.com",
					},
				},
				Summary: &MergeSummary{
					MatchCounts: map[string]int{
						"child1-domain.com": 2,
						"child2-domain.com": 1,
					},
					UnmappedCounts: map[string]int{
						"child1-domain.com": 0,
						"child2-domain.com": 1,
					},
				},
			},
			expectError: false,
		},
		// Add more test cases here
	}

	// マージ候補のセットを比較する補助関数
	compareCandidateSets := func(t *testing.T, got, expected []MergeCandidate) {
		t.Helper()
		if len(got) != len(expected) {
			t.Errorf("Candidate count mismatch: expected %d, got %d", len(expected), len(got))
			return
		}

		// 各期待値に対して、一致する候補が存在するか確認
		for _, exp := range expected {
			found := false
			for _, g := range got {
				if g.Parent.Email == exp.Parent.Email && g.Child.Email == exp.Child.Email {
					// 詳細な比較
					if g.Parent != exp.Parent {
						t.Errorf("Parent mismatch for %s\nExpected: %+v\nGot: %+v",
							exp.Child.Email, exp.Parent, g.Parent)
					}
					if g.Child != exp.Child {
						t.Errorf("Child mismatch for %s\nExpected: %+v\nGot: %+v",
							exp.Child.Email, exp.Child, g.Child)
					}
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Missing expected candidate: Parent=%s, Child=%s",
					exp.Parent.Email, exp.Child.Email)
			}
		}
	}

	// Unmappedのセットを比較する補助関数
	compareUnmappedSets := func(t *testing.T, got, expected []admina.Identity) {
		t.Helper()
		if len(got) != len(expected) {
			t.Errorf("Unmapped count mismatch: expected %d, got %d", len(expected), len(got))
			return
		}

		// メールアドレスをキーにしたマップを作成
		expectedMap := make(map[string]admina.Identity)
		for _, e := range expected {
			expectedMap[e.Email] = e
		}

		for _, g := range got {
			if e, exists := expectedMap[g.Email]; exists {
				if g != e {
					t.Errorf("Unmapped identity mismatch for %s\nExpected: %+v\nGot: %+v",
						g.Email, e, g)
				}
				delete(expectedMap, g.Email)
			} else {
				t.Errorf("Unexpected unmapped identity: %+v", g)
			}
		}

		for email := range expectedMap {
			t.Errorf("Missing expected unmapped identity: %s", email)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// テストケース固有の設定を使用
			result, err := findMergeCandidates(tc.identities, tc.config)

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result but got nil")
			}

			// 順序に依存しない検証
			compareCandidateSets(t, result.Candidates, tc.expectedResult.Candidates)
			compareUnmappedSets(t, result.Unmapped, tc.expectedResult.Unmapped)

			// サマリーの検証
			if tc.expectedResult.Summary != nil {
				// マッチ数の検証
				for domain, expectedCount := range tc.expectedResult.Summary.MatchCounts {
					if gotCount := result.Summary.MatchCounts[domain]; gotCount != expectedCount {
						t.Errorf("Match count mismatch for domain %s: expected %d, got %d",
							domain, expectedCount, gotCount)
					}
				}

				// アンマップ数の検証
				for domain, expectedCount := range tc.expectedResult.Summary.UnmappedCounts {
					if gotCount := result.Summary.UnmappedCounts[domain]; gotCount != expectedCount {
						t.Errorf("Unmapped count mismatch for domain %s: expected %d, got %d",
							domain, expectedCount, gotCount)
					}
				}
			}
		})
	}
}

// エラーケースのテスト
func TestMergeIdentitiesError(t *testing.T) {
	// ... エラーケースのテストを追加
}

// CSV出力のテスト
func TestPrintCSVMergeResult(t *testing.T) {
	// テスト用の一時ディレクトリを作成し、出力先として設定
	tempDir := t.TempDir()

	result := &MergeResult{
		Candidates: []MergeCandidate{
			{
				Parent: admina.Identity{
					ID:    "1",
					Email: "test@parent-domain.com",
				},
				Child: admina.Identity{
					ID:    "2",
					Email: "test@child-domain.com",
				},
			},
		},
		Unmapped: []admina.Identity{
			{
				ID:    "3",
				Email: "unmapped@child-domain.com",
			},
		},
	}

	// CSVFormatterを使用してCSV出力のテスト (マスクなし)
	csvFormatter := &CSVFormatter{OutputDir: tempDir}
	_, err := csvFormatter.Format(result, 1, 0, true) // noMask を true に設定
	assert.NoError(t, err)

	// 出力ファイルの存在確認
	mappingsPath := filepath.Join(tempDir, "identity_mappings.csv")
	unmappedPath := filepath.Join(tempDir, "unmapped_child_identities.csv")

	assert.FileExists(t, mappingsPath)
	assert.FileExists(t, unmappedPath)

	// ファイルの内容を検証 (マスクなし)
	mappingsContent, err := os.ReadFile(mappingsPath)
	assert.NoError(t, err)
	assert.Contains(t, string(mappingsContent), "test@parent-domain.com")

	unmappedContent, err := os.ReadFile(unmappedPath)
	assert.NoError(t, err)
	assert.Contains(t, string(unmappedContent), "unmapped@child-domain.com")

	// CSVFormatterを使用してCSV出力のテスト (マスクあり)
	_, err = csvFormatter.Format(result, 1, 0, false) // noMask を false に設定
	assert.NoError(t, err)

	// ファイルの内容を検証 (マスクあり)
	mappingsContent, err = os.ReadFile(mappingsPath)
	assert.NoError(t, err)
	assert.Contains(t, string(mappingsContent), "te**@parent-domain.com")

	unmappedContent, err = os.ReadFile(unmappedPath)
	assert.NoError(t, err)
	assert.Contains(t, string(unmappedContent), "un******@child-domain.com")
}
