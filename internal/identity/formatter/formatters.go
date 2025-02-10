package identity

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	"github.com/moneyforward-i/admina-sysutils/internal/identity"
)

// JSONFormatter の実装
type JSONFormatter struct{}

func (f *JSONFormatter) Format(result *identity.MergeResult, mergedCount, skippedCount int) (string, error) {
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

		data.Parent.Email = identity.MaskEmail(data.Parent.Email)
		data.Child.Email = identity.MaskEmail(data.Child.Email)

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

func (f *MarkdownFormatter) Format(result *identity.MergeResult, mergedCount, skippedCount int) (string, error) {
	var output strings.Builder

	output.WriteString("# Merge Result\n\n")
	output.WriteString("## Candidates\n\n")
	output.WriteString("| No. | Status | Parent | Child |\n")
	output.WriteString("|-----|--------|---------|--------|\n")

	for i, candidate := range result.Candidates {
		parentEmail := identity.MaskEmail(candidate.Parent.Email)
		childEmail := identity.MaskEmail(candidate.Child.Email)
		output.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
			i+1, candidate.Status, parentEmail, childEmail))
	}

	return output.String(), nil
}

// PrettyFormatter の実装
type PrettyFormatter struct{}

func (f *PrettyFormatter) Format(result *identity.MergeResult, mergedCount, skippedCount int) (string, error) {
	var output strings.Builder

	output.WriteString("=== Merge Result ===\n\n")
	output.WriteString("Candidates:\n")

	for i, candidate := range result.Candidates {
		parentEmail := identity.MaskEmail(candidate.Parent.Email)
		childEmail := identity.MaskEmail(candidate.Child.Email)
		output.WriteString(fmt.Sprintf("%d. %s -> %s\n",
			i+1, childEmail, parentEmail))
	}

	return output.String(), nil
}

// CSVFormatter の実装
type CSVFormatter struct {
	OutputDir string
}

func (f *CSVFormatter) Format(result *identity.MergeResult, mergedCount, skippedCount int) (string, error) {
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
		parentEmail := identity.MaskEmail(candidate.Parent.Email)
		childEmail := identity.MaskEmail(candidate.Child.Email)
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
		childEmail := identity.MaskEmail(unmapped.Email)
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
