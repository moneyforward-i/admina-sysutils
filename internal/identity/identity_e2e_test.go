package identity_test

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	"github.com/moneyforward-i/admina-sysutils/internal/identity"
	"github.com/moneyforward-i/admina-sysutils/internal/logger"
	"github.com/stretchr/testify/require"
)

const (
	excludeEmail = "murakami.katsutoshi@i.moneyforward.com"
)

// テストで作成したIdentityのIDを記録する
var (
	createdIdentityIDs = make(map[string]struct{})
	idMutex            sync.Mutex
	testRunID          string // テストの実行ID
	testLogFile        *os.File
)

// initTestLog はテストログファイルを初期化します
func initTestLog(t *testing.T) {
	// テストログ用のディレクトリを作成
	if err := os.MkdirAll("out/test", os.ModePerm); err != nil {
		t.Fatalf("Failed to create test log directory: %v", err)
	}

	var err error
	testLogFile, err = os.OpenFile("out/test/e2e_identity_test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}
}

// writeTestLog はテストログを出力します
func writeTestLog(format string, args ...interface{}) {
	if testLogFile != nil {
		fmt.Fprintf(testLogFile, format+"\n", args...)
	}
}

// generateTestRunID はテスト実行時の一意なIDを生成します
func generateTestRunID() string {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	return fmt.Sprintf("%06d", r.Intn(1000000))
}

// addTestRunID はメールアドレスにテストの実行IDを追加します
func addTestRunID(email string) string {
	return testRunID + "." + email
}

// addCreatedIdentityID は作成したIdentityのIDを記録します
func addCreatedIdentityID(id string) {
	idMutex.Lock()
	defer idMutex.Unlock()
	createdIdentityIDs[id] = struct{}{}
}

// verifyMergeResult はマージ結果を検証します
func verifyMergeResult(t *testing.T, ctx context.Context, client *admina.Client, parentEmail string, childEmail string) {
	// マージ完了を待機（セカンダリーメールアドレスの反映を待つ）
	time.Sleep(10 * time.Second)

	// parent-domainのIdentityを取得
	identities, err := identity.FetchAllIdentities(client)
	require.NoError(t, err, "Failed to fetch identities after merge")

	// 全てのIdentityの詳細をログ出力
	writeTestLog("\n=== All Identities After Merge ===")
	for i, identity := range identities {
		writeTestLog("Identity[%d]:", i)
		writeTestLog("  ID: %s", identity.ID)
		writeTestLog("  Email: %s", identity.Email)
		writeTestLog("  PeopleID: %d", identity.PeopleID)
		writeTestLog("  ManagementType: %s", identity.ManagementType)
		writeTestLog("  EmployeeType: %s", identity.EmployeeType)
		writeTestLog("  EmployeeStatus: %s", identity.EmployeeStatus)
		writeTestLog("  SecondaryEmails: %v", identity.SecondaryEmails)
	}
	writeTestLog("=== End of Identities ===\n")

	var parentIdentity admina.Identity
	var found bool
	for _, identity := range identities {
		if identity.Email == parentEmail {
			parentIdentity = identity
			found = true
			break
		}
	}

	require.True(t, found, "Parent identity should exist after merge")
	writeTestLog("Parent Identity after merge: %+v", parentIdentity)

	// マージ後の状態を検証
	require.Equal(t, "parent-domain.com", identity.ExtractDomain(parentIdentity.Email), "Parent email domain should be parent-domain.com")

	// セカンダリーメールアドレスの検証を強化
	writeTestLog("Verifying secondary emails. Expected child email: %s", childEmail)
	writeTestLog("Current secondary emails: %v", parentIdentity.SecondaryEmails)
	require.Contains(t, parentIdentity.SecondaryEmails, childEmail,
		"Child email should be included in secondary emails after merge")

	// マージ結果をログに記録
	writeTestLog("Merge Result:")
	writeTestLog("  From: %s", childEmail)
	writeTestLog("  To: %s", parentEmail)
	writeTestLog("  Status: Success")
	writeTestLog("  Secondary Emails: %v", parentIdentity.SecondaryEmails)
}

// TestE2E_Identity は実際の環境に対してE2Eテストを実行します
func TestE2E_Identity(t *testing.T) {
	// 環境変数のチェック
	if os.Getenv("E2E_TEST") != "1" {
		t.Skip("Skipping E2E test. Set E2E_TEST=1 to run this test")
	}

	// ロガーの初期化
	logger.Init()

	// テストログの初期化
	initTestLog(t)
	defer testLogFile.Close()

	// テストの実行IDを生成
	testRunID = generateTestRunID()
	logger.LogInfo("Generated test run ID: %s", testRunID)
	writeTestLog("=== Starting E2E Test with Run ID: %s ===", testRunID)

	// クライアントの初期化
	client := admina.NewClient()
	ctx := context.Background()

	if err := client.Validate(); err != nil {
		t.Fatal(err)
	}

	// テスト開始時に全てのIdentityを削除
	cleanupAllIdentities(t, ctx, client)
	// テスト終了時に全てのIdentityを削除
	defer cleanupAllIdentities(t, ctx, client)

	// テストケースの定義
	tests := []struct {
		name       string
		fromEmail  string
		toEmail    string
		shouldPass bool
	}{
		{
			name:       "External_to_External",
			fromEmail:  "tanaka.jiro.e2e.1@child2-ext-domain.com",
			toEmail:    "tanaka.jiro.e2e.2@child2-ext-domain.com",
			shouldPass: false, // 異なるローカルパートのため、マージされない
		},
		{
			name:       "External_to_Managed",
			fromEmail:  "suzuki.hanako.e2e@child2-ext-domain.com",
			toEmail:    "suzuki.hanako.e2e@parent-domain.com",
			shouldPass: true,
		},
		{
			name:       "Managed_to_Managed",
			fromEmail:  "yamada.taro.e2e@child1-domain.com",
			toEmail:    "yamada.taro.e2e@parent-domain.com",
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeTestLog("\n=== Test Case: %s ===", tt.name)

			// クリーンアップを実行
			cleanupAllIdentities(t, ctx, client)

			// fromIdentityの作成
			fromIdentity := findIdentityInCSV(t, tt.fromEmail)
			require.NotNil(t, fromIdentity, "Failed to find from identity in CSV")
			id, err := createIdentity(client, fromIdentity)
			require.NoError(t, err, "Failed to create from identity")
			addCreatedIdentityID(id)
			writeTestLog("Created From Identity: %+v", fromIdentity)

			// toIdentityの作成
			toIdentity := findIdentityInCSV(t, tt.toEmail)
			require.NotNil(t, toIdentity, "Failed to find to identity in CSV")
			id, err = createIdentity(client, toIdentity)
			require.NoError(t, err, "Failed to create to identity")
			addCreatedIdentityID(id)
			writeTestLog("Created To Identity: %+v", toIdentity)

			// マージ前のマトリックスを取得
			beforeMatrix, err := identity.GetIdentityMatrix(client)
			require.NoError(t, err, "Failed to get identity matrix before merge")
			writeTestLog("\nBefore Merge Matrix:\n%+v", beforeMatrix)

			// マージを実行
			err = identity.MergeIdentities(client, &identity.MergeConfig{
				ParentDomain: "parent-domain.com",
				ChildDomains: []string{"child1-domain.com", "child2-ext-domain.com"},
				DryRun:       false,
				AutoApprove:  true,
				OutputFormat: "json",
			})

			if tt.shouldPass {
				require.NoError(t, err, "Same merge should succeed")
				// マージ結果の詳細な検証
				verifyMergeResult(t, ctx, client, addTestRunID(tt.toEmail), addTestRunID(tt.fromEmail))
			} else {
				require.NoError(t, err, "Same merge should complete without error even if no merges occur")
			}

			// マージ後のマトリックスを取得
			afterMatrix, err := identity.GetIdentityMatrix(client)
			require.NoError(t, err, "Failed to get identity matrix after merge")
			writeTestLog("\nAfter Merge Matrix:\n%+v", afterMatrix)

			if tt.shouldPass {
				// マージが成功した場合の検証
				require.NotEqual(t, beforeMatrix, afterMatrix, "Matrix should change after successful merge")
				// マトリックスの期待値を検証
				if identity.ExtractDomain(tt.fromEmail) == "child1-domain.com" {
					// managed to managed のケース
					managedIndex := -1
					for i, mType := range afterMatrix.ManagementTypes {
						if mType == "managed" {
							managedIndex = i
							break
						}
					}
					require.NotEqual(t, -1, managedIndex, "Should have managed management type")
					require.Equal(t, 1, afterMatrix.Matrix[managedIndex][0], "Should have one managed to managed merge")
				} else if identity.ExtractDomain(tt.fromEmail) == "child2-ext-domain.com" {
					// external to managed のケース
					externalIndex := -1
					managedIndex := -1
					for i, mType := range afterMatrix.ManagementTypes {
						if mType == "external" {
							externalIndex = i
						} else if mType == "managed" {
							managedIndex = i
						}
					}
					require.NotEqual(t, -1, externalIndex, "Should have external management type")
					require.NotEqual(t, -1, managedIndex, "Should have managed management type")
					require.Equal(t, 1, afterMatrix.Matrix[externalIndex][0], "Should have one external to managed merge")
				}
			} else {
				// マージ対象外の場合の検証
				require.Equal(t, beforeMatrix, afterMatrix, "Matrix should not change when no merges occur")
				// マトリックスの値を検証
				for i := range afterMatrix.ManagementTypes {
					for j := range afterMatrix.Statuses {
						require.Equal(t, beforeMatrix.Matrix[i][j], afterMatrix.Matrix[i][j],
							fmt.Sprintf("Matrix value should not change at [%d][%d]", i, j))
					}
				}
			}

			// クリーンアップ
			cleanupAllIdentities(t, ctx, client)
		})
	}
}

// findIdentityInCSV はCSVファイルから指定されたメールアドレスのIdentityを探します
func findIdentityInCSV(t *testing.T, email string) *admina.CreateIdentityRequest {
	t.Helper()
	file, err := os.Open("testdata/e2e/identities.csv")
	if err != nil {
		t.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comment = '#'        // コメント行をスキップ
	reader.FieldsPerRecord = -1 // フィールド数の検証をスキップ

	headers, err := reader.Read()
	if err != nil {
		t.Fatalf("Failed to read CSV header: %v", err)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read CSV record: %v", err)
		}

		for i, header := range headers {
			if header == "primaryEmail" && record[i] == email {
				return createIdentityRequestFromRecord(headers, record)
			}
		}
	}
	return nil
}

func createIdentityRequestFromRecord(headers []string, record []string) *admina.CreateIdentityRequest {
	req := &admina.CreateIdentityRequest{}
	for i, header := range headers {
		if i >= len(record) {
			break
		}
		switch header {
		case "primaryEmail":
			req.PrimaryEmail = addTestRunID(record[i])
		case "employeeStatus":
			req.EmployeeStatus = record[i]
		case "firstName":
			req.FirstName = record[i]
		case "lastName":
			req.LastName = record[i]
		case "employeeType":
			req.EmployeeType = record[i]
		}
	}
	return req
}

// createIdentity はIdentityを作成し、作成完了を待機します
func createIdentity(client *admina.Client, req *admina.CreateIdentityRequest) (string, error) {
	logger.LogInfo("Creating identity with request: %+v", req)
	identity, err := client.CreateIdentity(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("failed to create identity: %w", err)
	}

	// 作成完了を待機
	time.Sleep(1 * time.Second)
	logger.LogInfo("Created identity: %s (ID: %s, Type: %s, Status: %s)",
		req.PrimaryEmail, identity.ID, req.EmployeeType, req.EmployeeStatus)
	return identity.ID, nil
}

// cleanupAllIdentities は全てのIdentityを削除します
func cleanupAllIdentities(t *testing.T, ctx context.Context, client *admina.Client) {
	t.Helper()

	// 全てのIdentityを取得
	identities, err := identity.FetchAllIdentities(client)
	if err != nil {
		t.Errorf("Failed to get identities: %v", err)
		return
	}

	// 全てのIdentityを削除
	for _, identity := range identities {
		if identity.Email == excludeEmail {
			continue
		}
		logger.LogInfo("Cleaning up identity: %s (%s)", identity.Email, identity.ID)
		err := client.DeleteIdentity(ctx, identity.ID)
		if err != nil {
			t.Errorf("Failed to delete identity %s: %v", identity.ID, err)
		}
		// APIレート制限を考慮して少し待機
		time.Sleep(500 * time.Millisecond)
	}

	// 削除完了を待機
	time.Sleep(2 * time.Second)

	// 削除が完了したことを確認
	maxRetries := 10 // リトライ回数を増やす
	for i := 0; i < maxRetries; i++ {
		remainingIdentities, err := identity.FetchAllIdentities(client)
		if err != nil {
			t.Errorf("Failed to get remaining identities: %v", err)
			return
		}

		// excludeEmail以外のIdentityが存在しないことを確認
		var remaining []string
		for _, identity := range remainingIdentities {
			if identity.Email != excludeEmail {
				remaining = append(remaining, identity.Email)
			}
		}

		if len(remaining) == 0 {
			logger.LogInfo("All identities have been cleaned up successfully")
			return
		}

		logger.LogInfo("Waiting for %d identities to be deleted: %v", len(remaining), remaining)
		time.Sleep(2 * time.Second) // 待機時間を延長
	}

	t.Fatalf("Failed to delete all identities after %d retries", maxRetries)
}
