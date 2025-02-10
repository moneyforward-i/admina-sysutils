package identity

import (
	"bufio"
	"context"
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

	result, err := prepareMergeResult(client, config)
	if err != nil {
		return err
	}

	mergedCount, skippedCount, errorCount := processMergeCandidates(ctx, client, config, result)

	if err := outputResults(result, config, mergedCount, skippedCount); err != nil {
		return err
	}

	if errorCount > 0 {
		return fmt.Errorf("completed with %d errors, %d merged, %d skipped", errorCount, mergedCount, skippedCount)
	}

	return nil
}

func prepareMergeResult(client Client, config *MergeConfig) (*MergeResult, error) {
	allIdentities, err := FetchAllIdentities(client)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch identities: %v", err)
	}

	result, err := findMergeCandidates(allIdentities, config)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func processMergeCandidates(ctx context.Context, client Client, config *MergeConfig, result *MergeResult) (mergedCount, skippedCount, errorCount int) {
	for i := range result.Candidates {
		candidate := &result.Candidates[i]
		status, err := processSingleCandidate(ctx, client, config, candidate)
		switch status {
		case "Success":
			mergedCount++
		case "Skip":
			skippedCount++
		case "Error":
			errorCount++
			candidate.Reason = fmt.Sprintf("Failed to merge: %v", err)
		}
	}
	return
}

func processSingleCandidate(ctx context.Context, client Client, config *MergeConfig, candidate *MergeCandidate) (string, error) {
	if !IsMergeAllowed(candidate.Parent, candidate.Child) {
		reason := fmt.Sprintf("cannot merge from %s to %s", candidate.Child.ManagementType, candidate.Parent.ManagementType)
		logger.LogInfo("%s (%s -> %s)", reason, MaskEmail(candidate.Child.Email), MaskEmail(candidate.Parent.Email))
		candidate.Status = "Skip"
		candidate.Reason = reason
		return "Skip", nil
	}

	if config.DryRun {
		logger.LogInfo("Dry-run: Would merge %s -> %s", MaskEmail(candidate.Child.Email), MaskEmail(candidate.Parent.Email))
		candidate.Status = "Skip"
		return "Skip", nil
	}

	if !config.AutoApprove {
		if !confirmMerge(candidate) {
			logger.LogInfo("Skipped merging %s -> %s", MaskEmail(candidate.Child.Email), MaskEmail(candidate.Parent.Email))
			candidate.Status = "Skip"
			return "Skip", nil
		}
	}

	clientMergeResult, err := client.MergeIdentities(ctx, candidate.Child.PeopleID, candidate.Parent.PeopleID)
	if err != nil {
		logger.LogInfo("Failed to merge %s -> %s: %v", MaskEmail(candidate.Child.Email), MaskEmail(candidate.Parent.Email), err)
		candidate.Status = "Error"
		return "Error", err
	}

	logger.LogInfo("Successfully merged %d -> %d", clientMergeResult.FromPeopleID, clientMergeResult.ToPeopleID)
	candidate.Status = "Success"
	return "Success", nil
}

func confirmMerge(candidate *MergeCandidate) bool {
	fmt.Printf("Merge %s -> %s? (y/n): ", MaskEmail(candidate.Child.Email), MaskEmail(candidate.Parent.Email))
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)
	return response == "y"
}

func outputResults(result *MergeResult, config *MergeConfig, mergedCount, skippedCount int) error {
	formatter, err := selectFormatter(config.OutputFormat)
	if err != nil {
		return err
	}

	output, err := formatter.Format(result, mergedCount, skippedCount)
	if err != nil {
		return fmt.Errorf("failed to format output: %v", err)
	}

	logger.LogInfo("Outputting merge results")
	logger.LogInfo("%s", output)

	logger.LogInfo("Writing CSV files")
	csvFormatter := &CSVFormatter{OutputDir: "out"}
	if _, err := csvFormatter.Format(result, mergedCount, skippedCount); err != nil {
		return fmt.Errorf("failed to write CSV files: %v", err)
	}

	return nil
}

func selectFormatter(format string) (Formatter, error) {
	switch format {
	case "json":
		return &JSONFormatter{}, nil
	case "markdown":
		return &MarkdownFormatter{}, nil
	case "pretty":
		return &PrettyFormatter{}, nil
	case "csv":
		return &CSVFormatter{OutputDir: "out"}, nil
	default:
		return nil, fmt.Errorf("unknown output format: %s", format)
	}
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
