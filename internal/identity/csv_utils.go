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
	isInit    bool
}

// NewCSVWriter は新しいCSVWriterを作成します
func NewCSVWriter(outputDir string) (*CSVWriter, error) {
	return &CSVWriter{
		outputDir: outputDir,
		isInit:    false,
	}, nil
}

// initOutputDir は出力ディレクトリを初期化します
func (w *CSVWriter) initOutputDir() error {
	if w.isInit {
		return nil
	}

	// 出力ディレクトリが存在する場合は削除
	if _, err := os.Stat(w.outputDir); err == nil {
		if err := os.RemoveAll(w.outputDir); err != nil {
			return fmt.Errorf("failed to remove existing directory: %v", err)
		}
	}

	// 出力ディレクトリの作成
	if err := os.MkdirAll(w.outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	w.isInit = true
	return nil
}

// WriteCSV はCSVファイルを書き込みます
func (w *CSVWriter) WriteCSV(filename string, headers []string, rows [][]string) error {
	// 初回書き込み時にディレクトリを初期化
	if err := w.initOutputDir(); err != nil {
		return err
	}

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
