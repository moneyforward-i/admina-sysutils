package identity

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
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
	NoMask       bool
}

type MergeCandidate struct {
	Parent admina.Identity
	Child  admina.Identity
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
	Format(result *MergeResult, mergedCount, skippedCount int, noMask bool) (string, error)
}

func MergeIdentities(client Client, config *MergeConfig) error {
	logger.LogInfo("Starting identity merge process")

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

	// マージ処理の実行
	for _, candidate := range result.Candidates {
		if config.DryRun {
			logger.LogInfo("Dry-run: Would merge %s -> %s", candidate.Child.Email, candidate.Parent.Email)
			skippedCount++
			continue
		}

		if !config.AutoApprove {
			fmt.Printf("Merge %s -> %s? (y/n): ", candidate.Child.Email, candidate.Parent.Email)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(response)
			if response != "y" {
				logger.LogInfo("Skipped merging %s -> %s", candidate.Child.Email, candidate.Parent.Email)
				skippedCount++
				continue
			}
		}

		if err := client.MergeIdentities(candidate.Parent.PeopleID, candidate.Child.PeopleID); err != nil {
			return fmt.Errorf("failed to merge identities: %v", err)
		}
		logger.LogInfo("Successfully merged %s -> %s", candidate.Child.Email, candidate.Parent.Email)
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
	output, err := formatter.Format(result, mergedCount, skippedCount, config.NoMask)
	if err != nil {
		return fmt.Errorf("failed to format output: %v", err)
	}

	// 結果の出力
	logger.LogInfo("Outputting merge results")
	fmt.Print(output)

	// CSVファイルの出力
	logger.LogInfo("Writing CSV files")
	csvFormatter := &CSVFormatter{OutputDir: "out"}
	if _, err := csvFormatter.Format(result, mergedCount, skippedCount, config.NoMask); err != nil {
		return fmt.Errorf("failed to write CSV files: %v", err)
	}

	// マージ予定の一覧とunmappedな一覧の出力
	printMergeAnalysis(result)

	return nil
}

func printMergeAnalysis(result *MergeResult) {
	logger.LogInfo("Printing merge analysis summary")
	fmt.Println("=== Merge Analysis Summary ===")
	fmt.Printf("Total merge candidates: %d\n", len(result.Candidates))
	fmt.Printf("Total unmapped identities: %d\n", len(result.Unmapped))
	fmt.Println("=== Analysis Complete ===")
}

// JSONFormatter の実装
type JSONFormatter struct{}

func (f *JSONFormatter) Format(result *MergeResult, mergedCount, skippedCount int, noMask bool) (string, error) {
	var output strings.Builder

	for i, candidate := range result.Candidates {
		data := struct {
			Index  int             `json:"index"`
			Status string          `json:"status"`
			Parent admina.Identity `json:"parent"`
			Child  admina.Identity `json:"child"`
		}{
			Index:  i + 1,
			Status: "Skipped",
			Parent: candidate.Parent,
			Child:  candidate.Child,
		}

		if !noMask {
			data.Parent.Email = MaskEmail(data.Parent.Email)
			data.Child.Email = MaskEmail(data.Child.Email)
		}

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

func (f *MarkdownFormatter) Format(result *MergeResult, mergedCount, skippedCount int, noMask bool) (string, error) {
	var output strings.Builder

	output.WriteString("# Merge Result\n\n")
	output.WriteString("## Candidates\n\n")
	output.WriteString("| No. | Status | Parent | Child |\n")
	output.WriteString("|-----|--------|---------|--------|\n")

	for i, candidate := range result.Candidates {
		parentEmail := candidate.Parent.Email
		childEmail := candidate.Child.Email
		if !noMask {
			parentEmail = MaskEmail(parentEmail)
			childEmail = MaskEmail(childEmail)
		}
		output.WriteString(fmt.Sprintf("| %d | Skipped | %s | %s |\n",
			i+1, parentEmail, childEmail))
	}

	return output.String(), nil
}

// PrettyFormatter の実装
type PrettyFormatter struct{}

func (f *PrettyFormatter) Format(result *MergeResult, mergedCount, skippedCount int, noMask bool) (string, error) {
	var output strings.Builder

	output.WriteString("=== Merge Result ===\n\n")
	output.WriteString("Candidates:\n")

	for i, candidate := range result.Candidates {
		parentEmail := candidate.Parent.Email
		childEmail := candidate.Child.Email
		if !noMask {
			parentEmail = MaskEmail(parentEmail)
			childEmail = MaskEmail(childEmail)
		}
		output.WriteString(fmt.Sprintf("%d. %s -> %s\n",
			i+1, childEmail, parentEmail))
	}

	return output.String(), nil
}

// CSVFormatter の実装
type CSVFormatter struct {
	OutputDir string
}

func (f *CSVFormatter) Format(result *MergeResult, mergedCount, skippedCount int, noMask bool) (string, error) {
	csvWriter, err := NewCSVWriter(f.OutputDir)
	if err != nil {
		return "", err
	}

	// マッピングファイルの作成
	mappingRows := make([][]string, 0, len(result.Candidates))
	for _, candidate := range result.Candidates {
		parentEmail := candidate.Parent.Email
		childEmail := candidate.Child.Email
		if !noMask {
			parentEmail = MaskEmail(parentEmail)
			childEmail = MaskEmail(childEmail)
		}
		mappingRows = append(mappingRows, []string{
			parentEmail,
			candidate.Parent.ID,
			childEmail,
			candidate.Child.ID,
			"Skipped",
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
		childEmail := unmapped.Email

		if !noMask {
			childEmail = MaskEmail(childEmail)
		}

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
	fmt.Printf("=== Starting Merge Analysis ===\n")
	fmt.Printf("Total identities to process: %d\n", len(identities))

	// 親ドメインのカウント
	parentCount := 0
	for _, identity := range identities {
		if ExtractDomain(identity.Email) == config.ParentDomain {
			parentCount++
		}
	}
	fmt.Printf("Parent domain (%s): %d identities\n", config.ParentDomain, parentCount)

	// 子ドメインのカウント
	childCounts := make(map[string]int)
	for _, domain := range config.ChildDomains {
		for _, identity := range identities {
			if ExtractDomain(identity.Email) == domain {
				childCounts[domain]++
			}
		}
	}
	fmt.Printf("Child domains:\n")
	for domain, count := range childCounts {
		fmt.Printf("  - %s: %d identities\n", domain, count)
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

	// 結果サマリーを出力
	fmt.Println("=== Merge Analysis Summary ===")
	fmt.Printf("Scanned identities: %d\n", len(identities))
	fmt.Printf("Parent domain (%s): %d identities\n", config.ParentDomain, parentCount)
	for domain, count := range childCounts {
		fmt.Printf("Child domain (%s): %d identities\n", domain, count)
		fmt.Printf("  - Matched: %d\n", result.Summary.MatchCounts[domain])
		fmt.Printf("  - Unmatched: %d\n", result.Summary.UnmappedCounts[domain])
	}
	fmt.Printf("Total merge candidates: %d\n", len(result.Candidates))
	fmt.Printf("Total unmapped identities: %d\n", len(result.Unmapped))
	fmt.Println("=== Analysis Complete ===")

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
