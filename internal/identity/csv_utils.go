package identity

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

// CSVWriter はCSVファイル作成のためのユーティリティ構造体です
type CSVWriter struct {
	outputDir string
}

// NewCSVWriter は新しいCSVWriterインスタンスを作成します
func NewCSVWriter(outputDir string) (*CSVWriter, error) {
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}
	return &CSVWriter{outputDir: outputDir}, nil
}

// WriteCSV はCSVファイルを作成し、データを書き込みます
func (w *CSVWriter) WriteCSV(filename string, headers []string, rows [][]string) error {
	filePath := filepath.Join(w.outputDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", filename, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %v", err)
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %v", err)
		}
	}

	return nil
}
