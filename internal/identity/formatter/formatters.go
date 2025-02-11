package identity

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	"github.com/moneyforward-i/admina-sysutils/internal/identity"
)

// ResultFormatter は結果のフォーマット方法を定義するインターフェース
type ResultFormatter interface {
	// Format は結果を文字列として返します
	Format(result *identity.MergeResult, mergedCount, skippedCount int) (string, error)
}

// CSVResultFormatter はCSV形式の結果を生成するインターフェース
type CSVResultFormatter interface {
	// FormatMappings はマッピング情報をCSV形式で返します
	FormatMappings(result *identity.MergeResult) ([][]string, error)
	// FormatUnmapped は未マッピング情報をCSV形式で返します
	FormatUnmapped(result *identity.MergeResult) ([][]string, error)
	// GetHeaders は各CSVのヘッダーを返します
	GetHeaders() (mappingHeaders, unmappedHeaders []string)
}

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
type CSVFormatter struct{}

func (f *CSVFormatter) Format(result *identity.MergeResult, mergedCount, skippedCount int) (string, error) {
	var output strings.Builder

	// マッピング情報の文字列化
	output.WriteString("=== Mapping Information ===\n")
	output.WriteString("ParentEmail,ParentIdentityID,ChildEmail,ChildIdentityID,Status\n")
	for _, candidate := range result.Candidates {
		parentEmail := identity.MaskEmail(candidate.Parent.Email)
		childEmail := identity.MaskEmail(candidate.Child.Email)
		output.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s\n",
			parentEmail,
			candidate.Parent.ID,
			childEmail,
			candidate.Child.ID,
			candidate.Status))
	}

	// アンマップ情報の文字列化
	output.WriteString("\n=== Unmapped Information ===\n")
	output.WriteString("ChildEmail,ChildIdentityID\n")
	for _, unmapped := range result.Unmapped {
		childEmail := identity.MaskEmail(unmapped.Email)
		output.WriteString(fmt.Sprintf("%s,%s\n",
			childEmail,
			unmapped.ID))
	}

	return output.String(), nil
}

func (f *CSVFormatter) FormatMappings(result *identity.MergeResult) ([][]string, error) {
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
	return mappingRows, nil
}

func (f *CSVFormatter) FormatUnmapped(result *identity.MergeResult) ([][]string, error) {
	unmappedRows := make([][]string, 0, len(result.Unmapped))
	for _, unmapped := range result.Unmapped {
		childEmail := identity.MaskEmail(unmapped.Email)
		unmappedRows = append(unmappedRows, []string{
			childEmail,
			unmapped.ID,
		})
	}
	return unmappedRows, nil
}

func (f *CSVFormatter) GetHeaders() (mappingHeaders, unmappedHeaders []string) {
	return []string{"ParentEmail", "ParentIdentityID", "ChildEmail", "ChildIdentityID", "Status"},
		[]string{"ChildEmail", "ChildIdentityID"}
}
