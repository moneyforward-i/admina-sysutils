package identity

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	"github.com/moneyforward-i/admina-sysutils/internal/logger"
)

type MergeConfig struct {
	ParentDomain string
	ChildDomains []string
	DryRun       bool
	AutoApprove  bool
	OutputFormat string
}

type MergeCandidate struct {
	Parent admina.Identity
	Child  admina.Identity
	Status string
	Reason string
}

type MergeSummary struct {
	TotalIdentities    int
	MergeCandidates    int
	UnmappedIdentities int
	MatchCounts        map[string]int
	UnmappedCounts     map[string]int
}

type MergeResult struct {
	Candidates []MergeCandidate
	Unmapped   []admina.Identity
	Summary    *MergeSummary
}

// Formatter はマージ結果のフォーマット方法を定義するインターフェース
type Formatter interface {
	Format(result *MergeResult, mergedCount, skippedCount int) (string, error)
}

func MergeIdentities(client Client, config *MergeConfig) error {
	logger.LogInfo("Starting identity merge process")
	ctx := context.Background()

	allIdentities, err := FetchAllIdentities(client)
	if err != nil {
		return fmt.Errorf("failed to fetch identities: %v", err)
	}

	result, err := findMergeCandidates(allIdentities, config)
	if err != nil {
		return err
	}

	mergedCount := 0
	skippedCount := 0
	errorCount := 0

	// マージ処理の実行
	for i, candidate := range result.Candidates {
		// マージ可否チェック
		if !IsMergeAllowed(candidate.Parent, candidate.Child) {
			reason := fmt.Sprintf("cannot merge from %s to %s", candidate.Child.ManagementType, candidate.Parent.ManagementType)
			logger.LogInfo("%s (%s -> %s)", reason, MaskEmail(candidate.Child.Email), MaskEmail(candidate.Parent.Email))
			result.Candidates[i].Status = "Skip" // インデックスを使用して直接更新
			result.Candidates[i].Reason = reason
			skippedCount++
			continue
		}

		if config.DryRun {
			logger.LogInfo("Dry-run: Would merge %s -> %s", MaskEmail(candidate.Child.Email), MaskEmail(candidate.Parent.Email))
			result.Candidates[i].Status = "Skip"
			skippedCount++
			continue
		}

		if !config.AutoApprove {
			fmt.Printf("Merge %s -> %s? (y/n): ", MaskEmail(candidate.Child.Email), MaskEmail(candidate.Parent.Email))
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(response)
			if response != "y" {
				logger.LogInfo("Skipped merging %s -> %s", MaskEmail(candidate.Child.Email), MaskEmail(candidate.Parent.Email))
				result.Candidates[i].Status = "Skip"
				skippedCount++
				continue
			}
		}

		clientMergeResult, err := client.MergeIdentities(ctx, candidate.Child.PeopleID, candidate.Parent.PeopleID)
		if err != nil {
			logger.LogInfo("Failed to merge %s -> %s: %v", MaskEmail(candidate.Child.Email), MaskEmail(candidate.Parent.Email), err)
			result.Candidates[i].Status = "Error"
			result.Candidates[i].Reason = fmt.Sprintf("Failed to merge: %v", err)
			errorCount++
			continue
		}
		logger.LogInfo("Successfully merged %d -> %d", clientMergeResult.FromPeopleID, clientMergeResult.ToPeopleID)
		result.Candidates[i].Status = "Success"
		mergedCount++
	}

	// フォーマッタの選択
	var formatter Formatter
	switch config.OutputFormat {
	case "json":
		formatter = &JSONFormatter{}
	case "markdown":
		formatter = &MarkdownFormatter{}
	case "pretty":
		formatter = &PrettyFormatter{}
	case "csv":
		formatter = &CSVFormatter{OutputDir: "out"}
	default:
		return fmt.Errorf("unknown output format: %s", config.OutputFormat)
	}

	// 結果のフォーマット
	output, err := formatter.Format(result, mergedCount, skippedCount)
	if err != nil {
		return fmt.Errorf("failed to format output: %v", err)
	}

	// 結果の出力（標準出力を使用）
	logger.LogInfo("Outputting merge results")
	logger.Print(output) // 実行結果のみ標準出力

	// CSVファイルの出力
	logger.LogInfo("Writing CSV files")
	csvFormatter := &CSVFormatter{OutputDir: "out"}
	if _, err := csvFormatter.Format(result, mergedCount, skippedCount); err != nil {
		return fmt.Errorf("failed to write CSV files: %v", err)
	}

	// エラーカウントがある場合は警告として返す
	if errorCount > 0 {
		return fmt.Errorf("completed with %d errors, %d merged, %d skipped", errorCount, mergedCount, skippedCount)
	}

	return nil
}

func printMergeAnalysis(result *MergeResult) {
	logger.LogInfo("Printing merge analysis summary")
	logger.PrintErr("=== Merge Analysis Summary ===\n")
	logger.PrintErr("Total merge candidates: %d\n", len(result.Candidates))
	logger.PrintErr("Total unmapped identities: %d\n", len(result.Unmapped))
	logger.PrintErr("=== Analysis Complete ===\n")
}

// JSONFormatter の実装
type JSONFormatter struct{}

func (f *JSONFormatter) Format(result *MergeResult, mergedCount, skippedCount int) (string, error) {
	var output strings.Builder

	for i, candidate := range result.Candidates {
		data := struct {
			Index  int             `json:"index"`
			Status string          `json:"status"`
			Parent admina.Identity `json:"parent"`
			Child  admina.Identity `json:"child"`
		}{
			Index:  i + 1,
			Status: candidate.Status,
			Parent: candidate.Parent,
			Child:  candidate.Child,
		}

		data.Parent.Email = MaskEmail(data.Parent.Email)
		data.Child.Email = MaskEmail(data.Child.Email)

		candidateJSON, err := json.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("failed to marshal candidate: %v", err)
		}
		output.WriteString(string(candidateJSON))
		output.WriteString("\n")
	}

	return output.String(), nil
}

// MarkdownFormatter の実装
type MarkdownFormatter struct{}

func (f *MarkdownFormatter) Format(result *MergeResult, mergedCount, skippedCount int) (string, error) {
	var output strings.Builder

	output.WriteString("# Merge Result\n\n")
	output.WriteString("## Candidates\n\n")
	output.WriteString("| No. | Status | Parent | Child |\n")
	output.WriteString("|-----|--------|---------|--------|\n")

	for i, candidate := range result.Candidates {
		parentEmail := MaskEmail(candidate.Parent.Email)
		childEmail := MaskEmail(candidate.Child.Email)
		output.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
			i+1, candidate.Status, parentEmail, childEmail))
	}

	return output.String(), nil
}

// PrettyFormatter の実装
type PrettyFormatter struct{}

func (f *PrettyFormatter) Format(result *MergeResult, mergedCount, skippedCount int) (string, error) {
	var output strings.Builder

	output.WriteString("=== Merge Result ===\n\n")
	output.WriteString("Candidates:\n")

	for i, candidate := range result.Candidates {
		parentEmail := MaskEmail(candidate.Parent.Email)
		childEmail := MaskEmail(candidate.Child.Email)
		output.WriteString(fmt.Sprintf("%d. %s -> %s\n",
			i+1, childEmail, parentEmail))
	}

	return output.String(), nil
}

// CSVFormatter の実装
type CSVFormatter struct {
	OutputDir string
}

func (f *CSVFormatter) Format(result *MergeResult, mergedCount, skippedCount int) (string, error) {
	// 環境変数が設定されているか確認
	projectRoot := os.Getenv("ADMINA_CLI_ROOT")
	if projectRoot == "" {
		// 環境変数が設定されていない場合、プロジェクトルートのディレクトリを取得
		var err error
		projectRoot, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %v", err)
		}
	}
	f.OutputDir = filepath.Join(projectRoot, "out", "data")

	csvWriter, err := NewCSVWriter(f.OutputDir)
	if err != nil {
		return "", err
	}

	// マッピングファイルの作成
	mappingRows := make([][]string, 0, len(result.Candidates))
	for _, candidate := range result.Candidates {
		parentEmail := MaskEmail(candidate.Parent.Email)
		childEmail := MaskEmail(candidate.Child.Email)
		mappingRows = append(mappingRows, []string{
			parentEmail,
			candidate.Parent.ID,
			childEmail,
			candidate.Child.ID,
			candidate.Status,
		})
	}

	if err := csvWriter.WriteCSV("identity_mappings.csv",
		[]string{"ParentEmail", "ParentIdentityID", "ChildEmail", "ChildIdentityID", "Status"},
		mappingRows); err != nil {
		return "", err
	}

	// アンマップファイルの作成
	unmappedRows := make([][]string, 0, len(result.Unmapped))
	for _, unmapped := range result.Unmapped {
		childEmail := MaskEmail(unmapped.Email)
		unmappedRows = append(unmappedRows, []string{
			childEmail,
			unmapped.ID,
		})
	}

	if err := csvWriter.WriteCSV("unmapped_child_identities.csv",
		[]string{"ChildEmail", "ChildIdentityID"},
		unmappedRows); err != nil {
		return "", err
	}

	return fmt.Sprintf("CSV files written to %s\n", f.OutputDir), nil
}

func findMergeCandidates(identities []admina.Identity, config *MergeConfig) (*MergeResult, error) {
	logger.PrintErr("=== Starting Merge Analysis ===\n")
	logger.PrintErr("Total identities to process: %d\n", len(identities))

	// 親ドメインのカウント
	parentCount := 0
	for _, identity := range identities {
		if ExtractDomain(identity.Email) == config.ParentDomain {
			parentCount++
		}
	}
	logger.PrintErr("Parent domain (%s): %d identities\n", config.ParentDomain, parentCount)

	// 子ドメインのカウント
	childCounts := make(map[string]int)
	for _, domain := range config.ChildDomains {
		for _, identity := range identities {
			if ExtractDomain(identity.Email) == domain {
				childCounts[domain]++
			}
		}
	}
	logger.PrintErr("Child domains:\n")
	for domain, count := range childCounts {
		logger.PrintErr("  - %s: %d identities\n", domain, count)
	}

	// マージ候補の検索
	result := &MergeResult{
		Candidates: []MergeCandidate{},
		Unmapped:   []admina.Identity{},
		Summary: &MergeSummary{
			TotalIdentities: len(identities),
			MatchCounts:     make(map[string]int),
			UnmappedCounts:  make(map[string]int),
		},
	}

	// マージ候補と未マッピングのカウント
	for _, identity := range identities {
		if contains(config.ChildDomains, ExtractDomain(identity.Email)) {
			localPart := ExtractLocalPart(identity.Email)
			matched := false
			for _, parent := range identities {
				if ExtractDomain(parent.Email) == config.ParentDomain &&
					ExtractLocalPart(parent.Email) == localPart {
					result.Candidates = append(result.Candidates, MergeCandidate{
						Parent: parent,
						Child:  identity,
					})
					result.Summary.MatchCounts[ExtractDomain(identity.Email)]++
					matched = true
					break
				}
			}
			if !matched {
				result.Unmapped = append(result.Unmapped, identity)
				result.Summary.UnmappedCounts[ExtractDomain(identity.Email)]++
			}
		}
	}

	// 結果サマリーの出力
	logger.PrintErr("=== Merge Analysis Summary ===\n")
	logger.PrintErr("Scanned identities: %d\n", len(identities))
	logger.PrintErr("Parent domain (%s): %d identities\n", config.ParentDomain, parentCount)
	for domain, count := range childCounts {
		logger.PrintErr("Child domain (%s): %d identities\n", domain, count)
		logger.PrintErr("  - Matched: %d\n", result.Summary.MatchCounts[domain])
		logger.PrintErr("  - Unmatched: %d\n", result.Summary.UnmappedCounts[domain])
	}
	logger.PrintErr("Total merge candidates: %d\n", len(result.Candidates))
	logger.PrintErr("Total unmapped identities: %d\n", len(result.Unmapped))
	logger.PrintErr("=== Analysis Complete ===\n")

	result.Summary.MergeCandidates = len(result.Candidates)
	result.Summary.UnmappedIdentities = len(result.Unmapped)

	return result, nil
}

// contains checks if a string exists in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// IDのタイプを確認する関数
func IsMergeAllowed(parent admina.Identity, child admina.Identity) bool {
	allowedMerges := map[string][]string{
		"managed":      {"managed"},
		"external":     {"managed", "external"},
		"system":       {"managed", "external", "system"},
		"unregistered": {"managed", "external", "system"},
		"unknown":      {"managed", "external", "system", "unregistered"},
	}

	childType := child.ManagementType
	parentType := parent.ManagementType

	for _, allowedType := range allowedMerges[childType] {
		if parentType == allowedType {
			return true
		}
	}
	return false
}
