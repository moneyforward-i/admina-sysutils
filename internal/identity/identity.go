package identity

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Identity struct {
	ManagementType string `json:"managementType"`
	EmployeeStatus string `json:"employeeStatus"`
	Status         string `json:"status"`
}

type Matrix struct {
	ManagementTypes []string
	Statuses        []string
	Matrix          [][]int
}

type Client interface {
	GetIdentities(cursor string) ([]Identity, string, int, error)
}

func GetIdentityMatrix(client Client) (*Matrix, error) {
	var allIdentities []Identity
	nextCursor := ""
	step := 0

	for {
		step++
		fmt.Fprintf(os.Stderr, "\rProcessing step: %d", step)

		identities, cursor, _, err := client.GetIdentities(nextCursor)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n")
			return nil, fmt.Errorf("failed to fetch identities: %v", err)
		}
		allIdentities = append(allIdentities, identities...)
		if cursor == "" {
			break
		}
		nextCursor = cursor
	}

	fmt.Fprintf(os.Stderr, "\nProcessing complete. Total steps: %d\n", step)
	fmt.Fprintf(os.Stderr, "Number of Identities retrieved: %d\n", len(allIdentities))

	if len(allIdentities) == 0 {
		return &Matrix{
			ManagementTypes: []string{},
			Statuses:        []string{},
			Matrix:          [][]int{},
		}, nil
	}

	return createMatrix(allIdentities)
}

func createMatrix(identities []Identity) (*Matrix, error) {
	matrix := &Matrix{
		ManagementTypes: []string{},
		Statuses:        []string{},
		Matrix:          [][]int{},
	}

	managementTypeMap := make(map[string]int)
	statusMap := make(map[string]int)

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

	matrix.Matrix = make([][]int, len(matrix.ManagementTypes))
	for i := range matrix.Matrix {
		matrix.Matrix[i] = make([]int, len(matrix.Statuses))
	}

	for _, identity := range identities {
		i := managementTypeMap[identity.ManagementType]
		j := statusMap[identity.EmployeeStatus]
		matrix.Matrix[i][j]++
	}

	orderedStatuses := []string{"draft", "preactive", "active", "on_leave", "retired", "untracked", "archived"}
	newStatuses := make([]string, 0, len(orderedStatuses))

	statusExists := make(map[string]bool)
	for _, status := range matrix.Statuses {
		statusExists[status] = true
	}

	for _, status := range orderedStatuses {
		if statusExists[status] {
			newStatuses = append(newStatuses, status)
		}
	}

	matrix.Statuses = newStatuses

	return matrix, nil
}

func PrintIdentityMatrix(client Client, outputFormat string) error {
	matrix, err := GetIdentityMatrix(client)
	if err != nil {
		return err
	}

	switch outputFormat {
	case "json":
		return printJSONMatrix(matrix)
	case "markdown":
		return printMarkdownMatrix(matrix)
	case "pretty":
		return printPrettyMatrix(matrix)
	default:
		return fmt.Errorf("unknown output format: %s", outputFormat)
	}
}

func printJSONMatrix(matrix *Matrix) error {
	jsonData, err := json.MarshalIndent(matrix, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}

func printMarkdownMatrix(matrix *Matrix) error {
	if len(matrix.ManagementTypes) == 0 || len(matrix.Statuses) == 0 {
		fmt.Println("# Identity Matrix")
		fmt.Println("No data available.")
		return nil
	}

	fmt.Println("# Identity Matrix")
	fmt.Print("|               |")
	for _, status := range matrix.Statuses {
		fmt.Printf(" %-15s |", status)
	}
	fmt.Println()

	fmt.Print("|---------------|")
	for range matrix.Statuses {
		fmt.Print("----------------|")
	}
	fmt.Println()

	totalIdentities := 0
	for i, managementType := range matrix.ManagementTypes {
		fmt.Printf("| %-13s |", managementType)
		rowTotal := 0
		for j := range matrix.Statuses {
			fmt.Printf(" %-15d |", matrix.Matrix[i][j])
			rowTotal += matrix.Matrix[i][j]
		}
		totalIdentities += rowTotal
		fmt.Println()
	}

	fmt.Print("| Total         |")
	for j := range matrix.Statuses {
		columnTotal := 0
		for i := range matrix.ManagementTypes {
			columnTotal += matrix.Matrix[i][j]
		}
		fmt.Printf(" %-15d |", columnTotal)
	}
	fmt.Printf("\n\nTotal Identities: %d\n", totalIdentities)

	return nil
}

func printPrettyMatrix(matrix *Matrix) error {
	fmt.Println("Identity Matrix:")
	fmt.Printf("%-20s", "")
	for _, status := range matrix.Statuses {
		fmt.Printf("%-15s", status)
	}
	fmt.Println()

	fmt.Println(strings.Repeat("-", 20+15*len(matrix.Statuses)))

	totalIdentities := 0
	for i, managementType := range matrix.ManagementTypes {
		fmt.Printf("%-20s", managementType)
		rowTotal := 0
		for j := range matrix.Statuses {
			fmt.Printf("%-15d", matrix.Matrix[i][j])
			rowTotal += matrix.Matrix[i][j]
		}
		totalIdentities += rowTotal
		fmt.Println()
	}

	fmt.Println(strings.Repeat("-", 20+15*len(matrix.Statuses)))

	fmt.Printf("%-20s", "Total")
	for j := range matrix.Statuses {
		columnTotal := 0
		for i := range matrix.ManagementTypes {
			columnTotal += matrix.Matrix[i][j]
		}
		fmt.Printf("%-15d", columnTotal)
	}
	fmt.Printf("\n\nTotal Identities: %d\n", totalIdentities)

	return nil
}
