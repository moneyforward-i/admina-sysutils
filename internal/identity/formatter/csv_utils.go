package identity

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

// CSVWriter はCSVファイルの書き込みを行うための構造体
type CSVWriter struct {
	outputDir string
}

// NewCSVWriter は新しいCSVWriterを作成します
func NewCSVWriter(outputDir string) (*CSVWriter, error) {
	// 出力ディレクトリの作成
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	return &CSVWriter{
		outputDir: outputDir,
	}, nil
}

// WriteCSV はCSVファイルを書き込みます
func (w *CSVWriter) WriteCSV(filename string, headers []string, rows [][]string) error {
	// ファイルパスの作成
	filePath := filepath.Join(w.outputDir, filename)

	// ファイルの作成
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// CSVライターの作成
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// ヘッダーの書き込み
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %v", err)
	}

	// データの書き込み
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %v", err)
		}
	}

	return nil
}
