package identity

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/moneyforward-i/admina-sysutils/internal/admina"
	"github.com/moneyforward-i/admina-sysutils/internal/logger"
)

type Matrix struct {
	ManagementTypes []string
	Statuses        []string
	Matrix          [][]int
}

// Formatter はマトリックス結果のフォーマット方法を定義するインターフェース
type MatrixFormatter interface {
	Format(matrix *Matrix) (string, error)
}

func GetIdentityMatrix(client Client) (*Matrix, error) {
	allIdentities, err := FetchAllIdentities(client)
	if err != nil {
		return nil, err
	}

	if len(allIdentities) == 0 {
		return &Matrix{
			ManagementTypes: []string{},
			Statuses:        []string{},
			Matrix:          [][]int{},
		}, nil
	}

	return createMatrix(allIdentities)
}

func PrintIdentityMatrix(client Client, outputFormat string) error {
	matrix, err := GetIdentityMatrix(client)
	if err != nil {
		return err
	}

	// フォーマッタの選択
	var formatter MatrixFormatter
	switch outputFormat {
	case "json":
		formatter = &JSONMatrixFormatter{}
	case "markdown":
		formatter = &MarkdownMatrixFormatter{}
	case "pretty":
		formatter = &PrettyMatrixFormatter{}
	default:
		return fmt.Errorf("unknown output format: %s", outputFormat)
	}

	// 結果のフォーマット
	output, err := formatter.Format(matrix)
	if err != nil {
		return fmt.Errorf("failed to format matrix: %v", err)
	}

	// 結果の出力
	logger.LogInfo("Outputting identity matrix")
	fmt.Print(output)
	return nil
}

// JSONMatrixFormatter の実装
type JSONMatrixFormatter struct{}

func (f *JSONMatrixFormatter) Format(matrix *Matrix) (string, error) {
	jsonData, err := json.MarshalIndent(matrix, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// MarkdownMatrixFormatter の実装
type MarkdownMatrixFormatter struct{}

func (f *MarkdownMatrixFormatter) Format(matrix *Matrix) (string, error) {
	if len(matrix.ManagementTypes) == 0 || len(matrix.Statuses) == 0 {
		return "# Identity Matrix\nNo data available.\n", nil
	}

	var output strings.Builder
	output.WriteString("# Identity Matrix\n")
	output.WriteString("| Type          |")
	for _, status := range matrix.Statuses {
		output.WriteString(fmt.Sprintf(" %-15s |", status))
	}
	output.WriteString(" Total          |\n")

	output.WriteString("|---------------|")
	for range matrix.Statuses {
		output.WriteString("----------------|")
	}
	output.WriteString("----------------|\n")

	totalIdentities := 0
	for i, managementType := range matrix.ManagementTypes {
		output.WriteString(fmt.Sprintf("| %-13s |", managementType))
		rowTotal := 0
		for j := range matrix.Statuses {
			output.WriteString(fmt.Sprintf(" %-15d |", matrix.Matrix[i][j]))
			rowTotal += matrix.Matrix[i][j]
		}
		output.WriteString(fmt.Sprintf(" %-15d |\n", rowTotal))
		totalIdentities += rowTotal
	}

	output.WriteString("| Total         |")
	for j := range matrix.Statuses {
		columnTotal := 0
		for i := range matrix.ManagementTypes {
			columnTotal += matrix.Matrix[i][j]
		}
		output.WriteString(fmt.Sprintf(" %-15d |", columnTotal))
	}
	output.WriteString(fmt.Sprintf(" %-15d |\n", totalIdentities))
	output.WriteString(fmt.Sprintf("\n\nTotal Identities: %d\n", totalIdentities))

	return output.String(), nil
}

// PrettyMatrixFormatter の実装
type PrettyMatrixFormatter struct{}

func (f *PrettyMatrixFormatter) Format(matrix *Matrix) (string, error) {
	var output strings.Builder
	output.WriteString("Identity Matrix:\n")
	output.WriteString(fmt.Sprintf("%-20s", "Type"))
	for _, status := range matrix.Statuses {
		output.WriteString(fmt.Sprintf("%-15s", status))
	}
	output.WriteString(fmt.Sprintf("%-15s\n", "Total"))

	output.WriteString(strings.Repeat("-", 20+15*len(matrix.Statuses)+15) + "\n")

	totalIdentities := 0
	for i, managementType := range matrix.ManagementTypes {
		output.WriteString(fmt.Sprintf("%-20s", managementType))
		rowTotal := 0
		for j := range matrix.Statuses {
			output.WriteString(fmt.Sprintf("%-15d", matrix.Matrix[i][j]))
			rowTotal += matrix.Matrix[i][j]
		}
		output.WriteString(fmt.Sprintf("%-15d\n", rowTotal))
		totalIdentities += rowTotal
	}

	output.WriteString(strings.Repeat("-", 20+15*len(matrix.Statuses)+15) + "\n")

	output.WriteString(fmt.Sprintf("%-20s", "Total"))
	for j := range matrix.Statuses {
		columnTotal := 0
		for i := range matrix.ManagementTypes {
			columnTotal += matrix.Matrix[i][j]
		}
		output.WriteString(fmt.Sprintf("%-15d", columnTotal))
	}
	output.WriteString(fmt.Sprintf("%-15d\n", totalIdentities))
	output.WriteString(fmt.Sprintf("\n\nTotal Identities: %d\n", totalIdentities))

	return output.String(), nil
}

// createMatrix を追加
func createMatrix(identities []admina.Identity) (*Matrix, error) {
	matrix := &Matrix{
		ManagementTypes: []string{},
		Statuses:        []string{},
		Matrix:          [][]int{},
	}

	managementTypeMap := make(map[string]int)
	statusMap := make(map[string]int)

	// 一意なManagementTypeとStatusを収集
	for _, identity := range identities {
		if _, exists := managementTypeMap[identity.ManagementType]; !exists {
			managementTypeMap[identity.ManagementType] = len(matrix.ManagementTypes)
			matrix.ManagementTypes = append(matrix.ManagementTypes, identity.ManagementType)
		}
		if _, exists := statusMap[identity.EmployeeStatus]; !exists {
			statusMap[identity.EmployeeStatus] = len(matrix.Statuses)
			matrix.Statuses = append(matrix.Statuses, identity.EmployeeStatus)
		}
	}

	// マトリックスの初期化
	matrix.Matrix = make([][]int, len(matrix.ManagementTypes))
	for i := range matrix.Matrix {
		matrix.Matrix[i] = make([]int, len(matrix.Statuses))
	}

	// カウントの集計
	for _, identity := range identities {
		i := managementTypeMap[identity.ManagementType]
		j := statusMap[identity.EmployeeStatus]
		matrix.Matrix[i][j]++
	}

	return matrix, nil
}
